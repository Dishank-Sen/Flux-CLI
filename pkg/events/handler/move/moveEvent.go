package move

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    roottimeline "exp1/internal/recorder/root-timeline"
    "exp1/internal/types"
    "exp1/pkg/interfaces"
)

type Move struct {
    OldPath string
    NewPath string
    Watcher interfaces.IWatcher
    Ctx     context.Context
}

func NewMove(ctx context.Context, oldPath, newPath string, watcher interfaces.IWatcher) *Move {
    return &Move{
        OldPath: filepath.Clean(oldPath),
        NewPath: filepath.Clean(newPath),
        Watcher: watcher,
        Ctx:     ctx,
    }
}

func (m *Move) Trigger() error {
    treePath := filepath.Join(".rec", "files", "fileTree.json")

    // load
    raw, err := os.ReadFile(treePath)
    if err != nil {
        return fmt.Errorf("read file tree: %w", err)
    }
    var tree types.FileTree
    if err := json.Unmarshal(raw, &tree); err != nil {
        return fmt.Errorf("unmarshal file tree: %w", err)
    }
    if tree.Files == nil {
        tree.Files = []*types.Node{}
    }

    // normalize paths stored in tree
    normalizeTreePaths(&tree)

    // remove the node by old path
    old := filepath.Clean(m.OldPath)
    new := filepath.Clean(m.NewPath)

    var removed *types.Node
    var ok bool
    tree.Files, removed, ok = removeNodeHelper(tree.Files, old)
    if !ok || removed == nil {
        // Not found: best-effort create node for new path (or return error)
        return fmt.Errorf("move: old node not found: %s", old)
    }

    // update removed subtree paths and name
    removed.Name = filepath.Base(new)
    updatePathsRecursive(removed, old, new)

    // insert under new parent
    newParent := filepath.Clean(filepath.Dir(new))
    if newParent == "." {
        newParent = ""
    }
    var errIns error
    tree.Files, errIns = insertNodeAtParentByPath(tree.Files, removed, newParent)
    if errIns != nil {
        return fmt.Errorf("move: insert failed: %w", errIns)
    }

    // save the tree atomically
    out, err := json.MarshalIndent(tree, "", "  ")
    if err != nil {
        return fmt.Errorf("marshal file tree: %w", err)
    }
    tmp := treePath + ".tmp"
    if err := os.WriteFile(tmp, out, 0o644); err != nil {
        return fmt.Errorf("write temp file tree: %w", err)
    }
    if err := os.Rename(tmp, treePath); err != nil {
        if err2 := os.WriteFile(treePath, out, 0o644); err2 != nil {
            return fmt.Errorf("final write file tree: %v (%w)", err, err2)
        }
    }

    // best-effort stat new path for metadata
    info, _ := os.Stat(new)

    // save timeline record (reuse types.Rename or create a types.Move)
    rec := types.Rename{
        NewPath:    new,
        NewName:    filepath.Base(new),
        Action:     "move",
        OldPath:    old,
        OldName:    filepath.Base(old),
        IsDir:      info != nil && info.IsDir(),
        Size:       0,
        RenameTime: time.Now(),
    }
    if info != nil {
        rec.Size = info.Size()
    }

    if err := roottimeline.Save(rec); err != nil {
        return fmt.Errorf("save timeline: %w", err)
    }

    return nil
}

// ---------- helper functions (copy/adapt from your rename code) ----------

func normalizeTreePaths(tree *types.FileTree) {
    var walk func(n *types.Node)
    walk = func(n *types.Node) {
        if n == nil {
            return
        }
        n.Path = filepath.Clean(n.Path)
        for _, c := range n.Children {
            walk(c)
        }
    }
    for _, f := range tree.Files {
        walk(f)
    }
}

func updatePathsRecursive(n *types.Node, oldPrefix, newPrefix string) {
    if n == nil {
        return
    }
    p := filepath.Clean(n.Path)
    rel, err := filepath.Rel(oldPrefix, p)
    if err != nil || rel == "." {
        n.Path = newPrefix
    } else {
        n.Path = filepath.Join(newPrefix, rel)
    }
    for _, c := range n.Children {
        updatePathsRecursive(c, oldPrefix, newPrefix)
    }
}

func insertNodeAtParentByPath(root []*types.Node, node *types.Node, parentPath string) ([]*types.Node, error) {
    parentPath = filepath.Clean(parentPath)
    if parentPath == "" || parentPath == "." || parentPath == string(filepath.Separator) {
        return append(root, node), nil
    }

    parentNode := findNodeByPath(root, parentPath)
    if parentNode != nil {
        parentNode.Children = append(parentNode.Children, node)
        return root, nil
    }

    parts := splitPathComponents(parentPath)
    if len(parts) == 0 {
        return append(root, node), nil
    }

    curParent := (*types.Node)(nil)
    curPath := ""
    for i, part := range parts {
        if curPath == "" {
            curPath = part
        } else {
            curPath = filepath.Join(curPath, part)
        }

        found := findNodeByPath(root, curPath)
        if found != nil {
            curParent = found
            continue
        }

        newDir := &types.Node{
            Name:     part,
            Path:     curPath,
            IsDir:    true,
            Children: []*types.Node{},
        }
        if curParent == nil {
            root = append(root, newDir)
        } else {
            curParent.Children = append(curParent.Children, newDir)
        }
        curParent = newDir

        if i == len(parts)-1 {
            curParent.Children = append(curParent.Children, node)
            return root, nil
        }
    }

    return append(root, node), nil
}

func splitPathComponents(p string) []string {
    clean := filepath.Clean(p)
    if clean == "." || clean == string(filepath.Separator) {
        return nil
    }
    parts := strings.Split(clean, string(filepath.Separator))
    var out []string
    for _, part := range parts {
        if part != "" {
            out = append(out, part)
        }
    }
    return out
}

func findNodeByPath(parent []*types.Node, target string) *types.Node {
    target = filepath.Clean(target)
    for _, n := range parent {
        if filepath.Clean(n.Path) == target {
            return n
        }
        if len(n.Children) > 0 {
            if found := findNodeByPath(n.Children, target); found != nil {
                return found
            }
        }
    }
    return nil
}

func removeNodeHelper(parent []*types.Node, target string) ([]*types.Node, *types.Node, bool) {
    target = filepath.Clean(target)
    for i := range parent {
        if filepath.Clean(parent[i].Path) == target {
            removed := parent[i]
            parent = append(parent[:i], parent[i+1:]...)
            return parent, removed, true
        }
        if len(parent[i].Children) > 0 {
            var removed *types.Node
            var ok bool
            parent[i].Children, removed, ok = removeNodeHelper(parent[i].Children, target)
            if ok {
                return parent, removed, true
            }
        }
    }
    return parent, nil, false
}
