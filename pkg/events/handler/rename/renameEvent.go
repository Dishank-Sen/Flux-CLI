package rename

import (
	"context"
	"encoding/json"
	roottimeline "exp1/internal/recorder/root-timeline"
	"exp1/internal/types"
	"exp1/pkg/interfaces"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Rename struct{
	OldPath string
	NewPath string
	Watcher interfaces.IWatcher
	Ctx context.Context
}

func NewRename(ctx context.Context, oldPath string, newpath string, watcher interfaces.IWatcher) *Rename{
	return &Rename{
		OldPath: oldPath,
		NewPath: newpath,
		Watcher: watcher,
		Ctx: ctx,
	}
}

func (r *Rename) Trigger() error {
	newName := filepath.Base(r.NewPath)
	oldName := filepath.Base(r.OldPath)

	// stat the new path to get metadata for timeline entry
	info, err := os.Stat(r.NewPath)
	if err != nil {
		return err
	}

	// 1) update file tree on disk
	treePath := filepath.Join(".rec", "files", "fileTree.json")
	data, err := os.ReadFile(treePath)
	if err != nil {
		return err
	}

	var tree types.FileTree
	if err := json.Unmarshal(data, &tree); err != nil {
		return err
	}
	if tree.Files == nil {
		tree.Files = []*types.Node{}
	}

	// attempt rename in tree
	renamed := false
	oldClean := filepath.Clean(r.OldPath)
	newClean := filepath.Clean(r.NewPath)

	for i := range tree.Files {
		tree.Files[i], renamed = findAndRename(tree.Files[i], oldClean, newClean, newName)
		if renamed {
			break
		}
	}

	if !renamed {
		// not found: return error or continue silently — choose to return error so caller knows
		return fmt.Errorf("rename: path not found in file tree: %s", r.OldPath)
	}

	out, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal file tree: %w", err)
	}
	if err := os.WriteFile(treePath, out, 0o644); err != nil {
		return fmt.Errorf("write file tree: %w", err)
	}

	// 2) prepare and save timeline/record of rename
	var dataRec = types.Rename{
		NewPath:    newClean,
		NewName:    newName,
		Action:     "rename",
		OldPath:    oldClean,
		OldName:    oldName,
		IsDir:      info.IsDir(),
		Size:       info.Size(),
		RenameTime: time.Now(),
	}

	return roottimeline.Save(dataRec)
}

func findAndRename(n *types.Node, oldPath, newPath, newName string) (*types.Node, bool) {
	if n == nil {
		return n, false
	}

	// match canonical paths
	nPath := filepath.Clean(n.Path)
	if nPath == oldPath {
		// update this node
		n.Name = newName
		n.Path = newPath

		// update all descendants paths by replacing oldPath prefix with newPath
		for _, c := range n.Children {
			updateDescendantPaths(c, oldPath, newPath)
		}
		return n, true
	}

	// recurse into children
	for i := range n.Children {
		updatedChild, ok := findAndRename(n.Children[i], oldPath, newPath, newName)
		if ok {
			n.Children[i] = updatedChild
			return n, true
		}
	}
	return n, false
}

// updateDescendantPaths replaces prefix oldPrefix in node.Path with newPrefix for node and descendants.
func updateDescendantPaths(n *types.Node, oldPrefix, newPrefix string) {
	if n == nil {
		return
	}
	p := filepath.Clean(n.Path)
	rel, err := filepath.Rel(oldPrefix, p)
	if err != nil || rel == "." {
		// if relative computation fails or it is the same path, set directly to newPrefix or join accordingly
		n.Path = newPrefix
	} else {
		// join newPrefix with the relative tail
		n.Path = filepath.Join(newPrefix, rel)
	}
	for _, c := range n.Children {
		updateDescendantPaths(c, oldPrefix, newPrefix)
	}
}