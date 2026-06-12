package utils

import (
	"context"
	"encoding/json"
	"errors"
	"exp1/internal/types"
	"exp1/utils"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func CreateDir(ctx context.Context, cancel context.CancelFunc, dirPath string, reinit bool) error{
    if reinit{
        if _, err := os.Stat(dirPath); err == nil {
            // directory already exists → skip
            return nil
        } else if !os.IsNotExist(err) {
            // unexpected error
            return err
        }

        // directory does not exist → create it
        return CreateDir(ctx, cancel, dirPath, false)
    }else{
        if err := os.MkdirAll(dirPath, 0o755); err != nil{
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
    path := path.Join(".rec", "config.json")
    if ctx.Err() != nil {
        cancel()
        return ctx.Err()
    }

    if reinit{
        if _, err := os.Stat(path); err == nil {
            // directory already exists → skip
            return nil
        } else if !os.IsNotExist(err) {
            // unexpected error
            return err
        }

        return CreateConfig(ctx, cancel, false)
    }else{
        // initial empty config
        repository := types.Repository{
            UserName: "",
            RemoteUrl: "",
        }
    
        recorder := types.Recorder{
            DebounceTime: 2,	//  initial default value
        }
    
        cfg := types.Config{
            Repository: repository,
            Recorder: recorder,
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

func CreateFileTree(ctx context.Context, cancel context.CancelFunc, reinit bool) error{
    fileTreePath := path.Join(".rec","files","fileTree.json")
    dirPath := path.Join(".rec", "files")

    if reinit{
        if info, err := os.Stat(fileTreePath); err == nil {
            // file already exists → skip
            if info.Size() == 0{
                // means file is empty so create a default file
                return CreateFileTree(ctx, cancel, false)
            }
            return nil
        } else if !os.IsNotExist(err) {
            // unexpected error
            return err
        }

        return CreateFileTree(ctx, cancel, false)
    }else{
        // check if the files folder exist
        exist := utils.CheckDirExist(dirPath)
        if !exist{
            // create files folder
            err := os.Mkdir(dirPath, 0755)
            if err != nil{
                return err
            }
        }
    
        f, err := os.OpenFile(fileTreePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
        if err != nil{
            return err
        }
        defer f.Close()
    
        var fileTree types.FileTree
        fileTree.Files = []*types.Node{}    // initially empty slice
    
        // write to fileTree.json
        data, err := json.MarshalIndent(fileTree, "", "  ")
        if err != nil{
            return err
        }
        
        if _, err := f.Write(data); err != nil {
            return err
        }
    }


    // check cancellation AFTER writing
    if ctx.Err() != nil {
        cancel()
        return errors.New("operation canceled during config creation")
    }

    return nil
}

func AddNode(cpath string) error {
    info, err := os.Stat(cpath)
    if err != nil {
        return err
    }

    node := &types.Node{
        Name:       filepath.Base(cpath),
        Path:       cpath,
        IsDir:      info.IsDir(),
        Size:       info.Size(),
        CreateTime: time.Now(),
    }

    components := getComponents(cpath)

    fileTreePath := filepath.Join(".rec", "files", "fileTree.json")
    byteData, err := os.ReadFile(fileTreePath)
    if err != nil {
        return err
    }

    var tree types.FileTree
    if err := json.Unmarshal(byteData, &tree); err != nil {
        return err
    }

    tree.Files = addNodeHelper(tree.Files, node, components)

    // marshal back and save
    out, err := json.MarshalIndent(tree, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(fileTreePath, out, 0644)
}


func getComponents(cpath string) []string{
    component := SplitPathComponents(cpath)
    if len(component) <= 1{
        return []string{}
    }
    return component[:len(component)-1]
}

func addNodeHelper(parent []*types.Node, newNode *types.Node, components []string) []*types.Node {
    if len(components) == 0 {
        return append(parent, newNode)
    }

    for i := range parent {
        if parent[i].Name == components[0] {
            parent[i].Children = addNodeHelper(parent[i].Children, newNode, components[1:])
            return parent
        }
    }

    return parent
}

func SplitPathComponents(p string) []string {
    clean := filepath.Clean(p)         // normalizes separators
    return strings.Split(clean, string(os.PathSeparator))
}

// func RemoveNode(cpath string) error{
//      node := &types.Node{
//         Name:       filepath.Base(cpath),
//         Path:       cpath,
//         IsDir:      info.IsDir(),
//         Size:       info.Size(),
//         CreateTime: time.Now(),
//     }

//     components := getComponents(cpath)

//     fileTreePath := filepath.Join(".rec", "files", "fileTree.json")
//     byteData, err := os.ReadFile(fileTreePath)
//     if err != nil {
//         return err
//     }

//     var tree types.FileTree
//     if err := json.Unmarshal(byteData, &tree); err != nil {
//         return err
//     }

//     tree.Files = addNodeHelper(tree.Files, node, components)

//     // marshal back and save
//     out, err := json.MarshalIndent(tree, "", "  ")
//     if err != nil {
//         return err
//     }

//     return os.WriteFile(fileTreePath, out, 0644)
// }

// func removeNodeHelper(parent []*types.Node, newNode *types.Node, components []string) []*types.Node {
//     if len(components) == 0 {
//         return append(parent, newNode)
//     }

//     for i := range parent {
//         if parent[i].Name == components[0] {
//             parent[i].Children = addNodeHelper(parent[i].Children, newNode, components[1:])
//             return parent
//         }
//     }

//     return parent
// }