package remove

import (
	"context"
	"encoding/json"
	roottimeline "exp1/internal/recorder/root-timeline"
	"exp1/internal/types"
	"exp1/pkg/interfaces"
	"exp1/utils/log"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Remove struct{
    Event fsnotify.Event
	Watcher interfaces.IWatcher
	Ctx context.Context
}

func NewRemove(ctx context.Context, event fsnotify.Event, watcher interfaces.IWatcher) *Remove{
	return &Remove{
		Event: event,
		Watcher: watcher,
		Ctx: ctx,
	}
}

func (r *Remove) Trigger() error{
	fileTreePath := path.Join(".rec", "files", "fileTree.json")
	removedNode, err := RemoveNodeFromFileTree(fileTreePath, r.Event.Name)
	if err != nil {
		return err
	}

	isDir := removedNode.IsDir
	size  := removedNode.Size
	ctime := removedNode.CreateTime

	path := r.Event.Name
    name := filepath.Base(path)
	msg := fmt.Sprintf("file removed: %s", path)
	log.Info(r.Ctx, msg)

	var data = types.Remove{
		Path: path,
        Name: name,
		Action: "remove",
        IsDir: isDir,
        Size: size,
        CreateTime: ctime,
		RemoveTime: time.Now(),
	}

	// add file to .rec/root-timeline
	return roottimeline.Save(data)
}

func removeNodeHelper(parent []*types.Node, target string) ([]*types.Node, *types.Node, bool) {
    target = filepath.Clean(target)

    for i := range parent {
        // Match full canonical path
        if filepath.Clean(parent[i].Path) == target {
            removed := parent[i]

            // Delete node i
            parent = append(parent[:i], parent[i+1:]...)

            return parent, removed, true
        }

        // Recurse into children
        if len(parent[i].Children) > 0 {
            var removed *types.Node
            var ok bool

            parent[i].Children, removed, ok =
                removeNodeHelper(parent[i].Children, target)

            if ok {
                return parent, removed, true
            }
        }
    }

    return parent, nil, false
}

func RemoveNodeFromFileTree(treePath, targetPath string) (*types.Node, error) {
    // Load the tree
    data, err := os.ReadFile(treePath)
    if err != nil {
        return nil, fmt.Errorf("read file tree: %w", err)
    }

    var tree types.FileTree
    if err := json.Unmarshal(data, &tree); err != nil {
        return nil, fmt.Errorf("unmarshal file tree: %w", err)
    }
    if tree.Files == nil {
        tree.Files = []*types.Node{}
    }

    // Remove node
    newFiles, removed, ok := removeNodeHelper(tree.Files, targetPath)
    if !ok {
        return nil, fmt.Errorf("path not found in tree: %s", targetPath)
    }
    tree.Files = newFiles

    // Save tree back to disk
    out, err := json.MarshalIndent(tree, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("marshal file tree: %w", err)
    }

    if err := os.WriteFile(treePath, out, 0o644); err != nil {
        return nil, fmt.Errorf("write file tree: %w", err)
    }

    return removed, nil
}
