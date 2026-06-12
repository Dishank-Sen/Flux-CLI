package initdir

import (
	"context"
	"sync"
)


type initDirfunc func(ctx context.Context, cancel context.CancelFunc, reinit bool) error
var InitDirectories []initDirfunc
var initDirMu sync.Mutex

func InitDir(f initDirfunc){
	if InitDirectories == nil{
		InitDirectories = make([]initDirfunc, 0)
	}
	initDirMu.Lock()
	InitDirectories = append(InitDirectories, f)
	initDirMu.Unlock()
}