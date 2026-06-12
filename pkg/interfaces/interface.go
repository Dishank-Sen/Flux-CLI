package interfaces

import (
	"context"
	"os"

	"github.com/fsnotify/fsnotify"
)

type IEvents interface{
	Create(event fsnotify.Event) error
	Remove(event fsnotify.Event) error
	Rename(event fsnotify.Event) error
	Write(event fsnotify.Event) error
	Flush() error
}

type IWatcher interface{
	AddDirToWatcher(ctx context.Context, path string, info os.FileInfo) error
}
