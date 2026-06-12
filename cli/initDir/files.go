package initdir

import (
	"context"
	"exp1/cli/utils"
	"path"
)

func init(){
	InitDir(CreateFiles)
}

func CreateFiles(ctx context.Context, cancel context.CancelFunc, reinit bool) error{
	dirPath := path.Join(".rec", "files")
	return utils.CreateDir(ctx, cancel, dirPath, reinit)
}