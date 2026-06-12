package watcher

import (
	"bufio"
	"context"
	"exp1/pkg/interfaces"
	"exp1/utils/log"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

var _ interfaces.IWatcher = (*Watch)(nil)

type Watch struct{
	watcher *fsnotify.Watcher
	events interfaces.IEvents
	Ctx context.Context
}

func NewWatcher(ctx context.Context) *Watch{
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		newCtx, cancel := context.WithCancel(ctx)
		log.Error(newCtx, cancel, err.Error())
		return nil
	}

	return &Watch{
		watcher: watcher,
		events: nil,
		Ctx: ctx,
	}
}

func (w *Watch) SetEvents(e interfaces.IEvents) {
    w.events = e
}

func (w *Watch) Start(ctx context.Context) error{
	err := w.filterFiles("./")
	if err != nil{
		return err
	}
	// here code will be blocked
	err = w.eventLoop(ctx)
	if err != nil{
		return err
	}
	return nil
}

// this loop never terminates until user explicitly turns it off

func (w *Watch) eventLoop(ctx context.Context) error {
	if w == nil || w.watcher == nil {
		return fmt.Errorf("watcher not initialized")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-w.watcher.Events:
			if !ok {
				// events channel closed by fsnotify; treat as clean shutdown
				return nil
			}

			patterns, err := w.getIgnoredFiles()
			if err != nil{
				return err
			}
			if w.matchesIgnore(event.Name, patterns) {
				continue
			}

			// route to handlers; use != 0 to test bitflags
			if event.Op&fsnotify.Create != 0 {
				if err := w.safeCallCreate(event); err != nil {
					return err
				}
			}
			if event.Op&fsnotify.Write != 0 {
				if err := w.safeCallWrite(event); err != nil {
					return err
				}
			}
			if event.Op&fsnotify.Remove != 0 {
				if err := w.safeCallRemove(event); err != nil {
					return err
				}
			}
			if event.Op&fsnotify.Rename != 0 {
				if err := w.safeCallRename(event); err != nil {
					return err
				}
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				// errors channel closed: treat as clean shutdown
				return nil
			}
			// propagate watcher error to caller
			return fmt.Errorf("fsnotify error: %w", err)
		}
	}
}

// helper wrappers that avoid nil deref and let you convert handler panics to errors if desired
func (w *Watch) safeCallCreate(ev fsnotify.Event) error {
	if w.events == nil {
		// decide policy: return error so caller can log/exit, or ignore
		return fmt.Errorf("events handler is nil")
	}
	w.events.Create(ev)
	return nil
}
func (w *Watch) safeCallWrite(ev fsnotify.Event) error {
	if w.events == nil {
		return fmt.Errorf("events handler is nil")
	}
	w.events.Write(ev)
	return nil
}
func (w *Watch) safeCallRemove(ev fsnotify.Event) error {
	if w.events == nil {
		return fmt.Errorf("events handler is nil")
	}
	w.events.Remove(ev)
	return nil
}
func (w *Watch) safeCallRename(ev fsnotify.Event) error {
	if w.events == nil {
		return fmt.Errorf("events handler is nil")
	}
	w.events.Rename(ev)
	return nil
}

// this removes all files mentioned in .recignore

func (w *Watch) filterFiles(root string) error {
    ignoredPatterns, err := w.getIgnoredFiles()
	if err != nil{
		return err
	}

    return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // If directory matches ignore pattern, skip it entirely
        if info.IsDir() && w.matchesIgnore(path, ignoredPatterns) {
            return filepath.SkipDir
        }

        // Otherwise, add the directory
        return w.AddDirToWatcher(w.Ctx, path, info)
    })
}

func (w *Watch) getIgnoredFiles() ([]string, error){
	ignoredPatterns, err := w.loadIgnore(filepath.Join("./", ".recignore"))
	if err != nil && !os.IsNotExist(err) { 
        // ignore error if .recignore not found
		// fmt.Println(err)
		return []string{},err
	}
	return ignoredPatterns, nil
}

func (w *Watch) loadIgnore(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}
	return patterns, scanner.Err()
}

func (w *Watch) matchesIgnore(path string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}
		// Handle directory patterns like "vendor/"
		if strings.HasSuffix(pattern, "/") && strings.Contains(path, strings.TrimSuffix(pattern, "/")) {
			return true
		}
	}
	return false
}

// add a dir to be watched

func (w *Watch) AddDirToWatcher(ctx context.Context, path string, info os.FileInfo) error{
	// Add directories to watcher
	if info.IsDir() {
		msg := fmt.Sprintf("watching: %s", path)
		log.Info(ctx, msg)
		return w.watcher.Add(path)
	}
	return nil
}