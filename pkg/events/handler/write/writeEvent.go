package write

import (
	"context"
	savehistory "exp1/internal/recorder/saveHistory"
	"exp1/internal/types"
	"exp1/pkg/interfaces"
	"exp1/utils/log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type Write struct {
	Event   fsnotify.Event
	Watcher interfaces.IWatcher
	State   map[string]types.Write
	Unsaved map[string]bool
	Ctx context.Context
}

func NewWrite(ctx context.Context, event fsnotify.Event, watcher interfaces.IWatcher, state map[string]types.Write, unsaved map[string]bool) *Write {
	return &Write{
		Event:   event,
		Watcher: watcher,
		State:   state,
		Unsaved: unsaved,
		Ctx: ctx,
	}
}

func (w *Write) Trigger() error{
	path := w.Event.Name
	ctx := w.Ctx

	info, err := os.Stat(path)
	if err != nil {
		// log.Error(ctx, cancel, err.Error())
		return err
	}
	size := info.Size()

	// Read new content
	newContentBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	newContent := string(newContentBytes)

	record, exists := w.State[path]

	// ===== Case 1: First time seeing this file =====
	if !exists {
		log.Info(ctx, "New file detected, creating snapshot record")

		data := types.Write{
			Path:                path,
			Type:                "snapshot",
			PrevSize:            size,
			CurrentSize:         size,
			PreviousFileContent: newContent,
			Timestamp:           time.Now(),
		}
		w.State[path] = data

		// save to history
		historyData := types.Write{
			Path: path,
			Type: "snapshot",
			Content: newContent,
			Timestamp: time.Now(),
			CurrentSize: size,
		}
		
		if err := savehistory.Save(historyData); err != nil{
			return err
		}
		// fmt.Println("history created for write snpashot!")
		
		w.Unsaved[path] = false

		return nil
	}

	// if file already tracked
	record.CurrentSize = size
	// fmt.Println("Current size:", record.CurrentSize)
	// fmt.Println("Previous size:", record.PrevSize)

	threshold, err := strconv.Atoi(os.Getenv("CODE_THRESHOLD"))
	if err != nil {
		// log.Fatal("Invalid CODE_THRESHOLD:", err)
		return err
	}

	// Only compute diff if file changed significantly
	if math.Abs(float64(record.CurrentSize)-float64(record.PrevSize)) > float64(threshold) {
		log.Info(ctx, "Significant change detected!")

		currentContentByte, err := os.ReadFile(path)
		if err != nil{
			// log.Fatal("error reading file (writeFile.go):",err)
			return err
		}

		currentContent := string(currentContentByte)
		previousContent := record.PreviousFileContent

		// create patch
		dmp := diffmatchpatch.New()
		patch := dmp.PatchMake(previousContent, currentContent)

		patchText := dmp.PatchToText(patch)

		// save history
		historyData := types.Write{
			Path: path,
			Type: "delta",
			Content: patchText,
			Timestamp: time.Now(),
			CurrentSize: size,
		}

		err = savehistory.Save(historyData)
		if err != nil{
			// log.Fatal("error occured (writeEvent.go):",err)
			return err
		}

		log.Info(ctx, "history created for write delta!")

		record.CurrentSize = size
		record.PrevSize = size
		record.PreviousFileContent = currentContent
		w.State[path] = record
		w.Unsaved[path] = false

		return nil
	} else {
		// fmt.Println("No significant change detected")
		w.Unsaved[path] = true
		return nil
	}
}

func (w *Write) Flush() error{
	// save snapshot file for every unsaved changes
	var unsavedFiles []string
	for key, value := range w.Unsaved{
		if value{
			unsavedFiles = append(unsavedFiles, key)
		}
	}

	for _, path := range unsavedFiles{
		content, err := os.ReadFile(path)
		if err != nil{
			// log.Fatal("error occured (writeEvent.go):", err)
			return err
		}

		info, err := os.Stat(path)
		if err != nil{
			return err
		}
		size := info.Size()
		stringContent := string(content)

		historyData := types.Write{
			Path: path,
			Type: "snapshot",
			Content: stringContent,
			Timestamp: time.Now(),
			CurrentSize: size,
		}
		if err = savehistory.Save(historyData); err != nil{
			return err
		}
	}
	return nil
}