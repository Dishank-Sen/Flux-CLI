package events

import (
	"context"
	"exp1/cli/utils"
	"exp1/internal/debounce"
	"exp1/internal/types"
	"exp1/pkg/events/handler/create"
	"exp1/pkg/events/handler/move"
	"exp1/pkg/events/handler/remove"
	"exp1/pkg/events/handler/rename"
	"exp1/pkg/events/handler/write"
	"exp1/pkg/interfaces"
	"exp1/utils/log"
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Events struct{
	watcher interfaces.IWatcher
	debouncer *debounce.Debouncer
	State map[string]types.Write
	Unsaved map[string]bool
	RenameFile map[string]time.Time
	Ctx context.Context
}

func NewEvents(w interfaces.IWatcher, ctx context.Context) *Events{
	return &Events{
		watcher: w,
		debouncer: debounce.NewDebouncer(),
		State: make(map[string]types.Write),
		Unsaved: make(map[string]bool),
		RenameFile: make(map[string]time.Time),
		Ctx: ctx,
	}
}

func (e *Events) Create(event fsnotify.Event) error {
    newPath := filepath.Clean(event.Name)
    newBase := filepath.Base(newPath)

    // 1) Check recent rename candidates (within window)
    // Try to find an oldPath candidate whose basename matches newBase.
    // Choose the most-recent candidate.
    var bestOld string
    var bestT time.Time
    for oldPath, t := range e.RenameFile {
        if time.Since(t) >= 1*time.Second {
            // expired candidate
            continue
        }
        if filepath.Base(filepath.Clean(oldPath)) == newBase {
            if bestOld == "" || t.After(bestT) {
                bestOld = filepath.Clean(oldPath)
                bestT = t
            }
        }
    }

    if bestOld != "" {
        // found a likely rename/move candidate
        log.Info(e.Ctx, fmt.Sprintf("Detected rename/move: %s → %s", bestOld, newPath))

        // consume candidate
        delete(e.RenameFile, bestOld)

        // decide move vs rename: if parent dir changed => move; else rename
        oldParent := filepath.Clean(filepath.Dir(bestOld))
        newParent := filepath.Clean(filepath.Dir(newPath))

        if oldParent != newParent {
            // it's a move across directories
            mv := move.NewMove(e.Ctx, bestOld, newPath, e.watcher)
            return mv.Trigger()
        } else {
            // rename within same directory (name changed)
            rn := rename.NewRename(e.Ctx, bestOld, newPath, e.watcher)
            return rn.Trigger()
        }
    }

    // 2) Not a move/rename -> proceed with normal create
    // AddNode should be idempotent (skip if exists)
    if err := utils.AddNode(newPath); err != nil {
        return err
    }

    createHandler := create.NewCreate(e.Ctx, event, e.watcher)
    return createHandler.Trigger()
}

func (e *Events) Remove(event fsnotify.Event) error{
	// fmt.Println("remove event:",event)
	remove := remove.NewRemove(e.Ctx, event, e.watcher)
	return remove.Trigger()
}

func (e *Events) Rename(event fsnotify.Event) error{
	// fmt.Println("rename event:",event)
	e.RenameFile[event.Name] = time.Now()
	return nil
}

func (e *Events) Write(event fsnotify.Event) error{
	path := event.Name
	var err error
	// Debounce per file path
	debounceTime, err := debounce.GetDebounceTime()
	if err != nil{
		return err
	}
	if debounceTime == 0{
		log.Info(e.Ctx, "no debounce time set")
		debounce.SetDebounceTime(2)
		debounceTime, err = debounce.GetDebounceTime()
		if err != nil{
			return err
		}
	}

	e.debouncer.Debounce(path, time.Duration(debounceTime)*time.Second, func() {
		writeHandler := write.NewWrite(e.Ctx, event, e.watcher, e.State, e.Unsaved)
		err = writeHandler.Trigger()
	})
	return err
}

func (e *Events) Flush() error{
	event := fsnotify.Event{}  // empty event
	writeHandler := write.NewWrite(e.Ctx, event, e.watcher, e.State, e.Unsaved)
	return writeHandler.Flush()
}