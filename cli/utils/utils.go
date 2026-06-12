package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Dishank-Sen/Flux-CLI/types"
	"github.com/Dishank-Sen/Flux-CLI/utils"
)

func CreateDir(ctx context.Context, cancel context.CancelFunc, dirPath string, reinit bool) error {
	if reinit {
		if _, err := os.Stat(dirPath); err == nil {
			// directory already exists → skip
			return nil
		} else if !os.IsNotExist(err) {
			// unexpected error
			return err
		}

		// directory does not exist → create it
		return CreateDir(ctx, cancel, dirPath, false)
	} else {
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			return err
		}
	}

	if ctx.Err() != nil {
		cancel()
		return errors.New("operation canceled during config creation")
	}
	return nil
}

func CreateConfig(ctx context.Context, cancel context.CancelFunc, reinit bool) error {
	path := path.Join(".flux", "config.json")
	if ctx.Err() != nil {
		cancel()
		return ctx.Err()
	}

	if reinit {
		if _, err := os.Stat(path); err == nil {
			// directory already exists → skip
			return nil
		} else if !os.IsNotExist(err) {
			// unexpected error
			return err
		}

		return CreateConfig(ctx, cancel, false)
	} else {
		// initial empty config
		repository := types.Repository{
			UserName:  "",
			RemoteUrl: "",
		}

		recorder := types.Recorder{
			DebounceTime:  3, //  initial default value
			CodeThreshold: 10,
		}

		sshKeys := types.SSHKeys{
			PublicKeyPath:  "",
			PrivateKeyPath: "",
		}

		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		cfg := types.Config{
			WorkingDir: dir,
			Repository: repository,
			Recorder:   recorder,
			SSHKeys:    sshKeys,
		}

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		// create file with read and write permission for owner only (0644)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
		defer f.Close()

		if _, err := f.Write(data); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
	}

	// check cancellation AFTER writing
	if ctx.Err() != nil {
		cancel()
		return errors.New("operation canceled during config creation")
	}

	return nil
}

var strictIgnore = map[string]struct{}{
	".flux":        {},
	".git":         {},
	"node_modules": {},
}

func CreateFileTree(ctx context.Context) error {
	fileTreePath := filepath.Join(".flux", "files", "fileTree.json")
	dirPath := filepath.Join(".flux", "files")

	if !utils.CheckDirExist(dirPath) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}
	}

	ignoreSet, err := loadIgnorePaths(".flowignore")
	if err != nil {
		return err
	}

	rootNodes := []*types.Node{}

	cfg, err := utils.GetConfig()
	if err != nil {
		return err
	}

	wd := cfg.WorkingDir

	entries, err := os.ReadDir(wd)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()

		if shouldIgnore(name, ignoreSet) {
			continue
		}

		fullPath := filepath.Join(wd, name)

		node, err := buildNode(wd, fullPath, ignoreSet)
		if err != nil {
			return err
		}

		rootNodes = append(rootNodes, node)
	}

	fileTree := types.FileTree{
		Files: rootNodes,
	}

	data, err := json.MarshalIndent(fileTree, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fileTreePath, data, 0644)
}

func buildNode(root string, path string, ignoreSet map[string]struct{}) (*types.Node, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return nil, err
	}

	node := &types.Node{
		Name:  info.Name(),
		Path:  relPath,
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		node.Size = info.Size()
		return node, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	children := []*types.Node{}

	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())

		if shouldIgnore(entry.Name(), ignoreSet) {
			continue
		}

		childNode, err := buildNode(root, childPath, ignoreSet)
		if err != nil {
			return nil, err
		}

		children = append(children, childNode)
	}

	if len(children) > 0 {
		node.Children = children
	}

	return node, nil
}

func loadIgnorePaths(ignoreFile string) (map[string]struct{}, error) {
	ignoreSet := make(map[string]struct{})

	data, err := os.ReadFile(ignoreFile)
	if err != nil {
		if os.IsNotExist(err) {
			return ignoreSet, nil
		}
		return nil, err
	}

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// normalize path
		line = filepath.Clean(line)

		ignoreSet[line] = struct{}{}
	}

	return ignoreSet, nil
}

func shouldIgnore(path string, ignoreSet map[string]struct{}) bool {
	path = filepath.Clean(path)
	base := filepath.Base(path)

	// strict ignore (always ignored)
	if _, ok := strictIgnore[base]; ok {
		return true
	}

	for ignore := range ignoreSet {
		if path == ignore || strings.HasPrefix(path, ignore+string(os.PathSeparator)) {
			return true
		}
	}

	return false
}

// func AddNode(cpath string) error {
// 	info, err := os.Stat(cpath)
// 	if err != nil {
// 		return err
// 	}

// 	node := &types.Node{
// 		Name:       filepath.Base(cpath),
// 		Path:       cpath,
// 		IsDir:      info.IsDir(),
// 		Size:       info.Size(),
// 		CreateTime: time.Now(),
// 	}

// 	components := getComponents(cpath)

// 	fileTreePath := filepath.Join(".flux", "files", "fileTree.json")
// 	byteData, err := os.ReadFile(fileTreePath)
// 	if err != nil {
// 		return err
// 	}

// 	var tree types.FileTree
// 	if err := json.Unmarshal(byteData, &tree); err != nil {
// 		return err
// 	}

// 	tree.Files = addNodeHelper(tree.Files, node, components)

// 	// marshal back and save
// 	out, err := json.MarshalIndent(tree, "", "  ")
// 	if err != nil {
// 		return err
// 	}

// 	return os.WriteFile(fileTreePath, out, 0644)
// }

// func getComponents(cpath string) []string {
// 	component := SplitPathComponents(cpath)
// 	if len(component) <= 1 {
// 		return []string{}
// 	}
// 	return component[:len(component)-1]
// }

// func addNodeHelper(parent []*types.Node, newNode *types.Node, components []string) []*types.Node {
// 	if len(components) == 0 {
// 		return append(parent, newNode)
// 	}

// 	for i := range parent {
// 		if parent[i].Name == components[0] {
// 			parent[i].Children = addNodeHelper(parent[i].Children, newNode, components[1:])
// 			return parent
// 		}
// 	}

// 	return parent
// }

// func SplitPathComponents(p string) []string {
// 	clean := filepath.Clean(p) // normalizes separators
// 	return strings.Split(clean, string(os.PathSeparator))
// }
