package initfiles

import (
	"context"
	"sync"
)


type initFilefunc func(ctx context.Context, cancel context.CancelFunc, reinit bool) error
var InitFiles []initFilefunc
var initFilesMu sync.Mutex

func InitFile(f initFilefunc){
	if InitFiles == nil{
		InitFiles = make([]initFilefunc, 0)
	}
	initFilesMu.Lock()
	InitFiles = append(InitFiles, f)
	initFilesMu.Unlock()
}