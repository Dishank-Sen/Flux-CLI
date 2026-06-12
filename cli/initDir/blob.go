package initdir

import (
	"context"
	"exp1/cli/utils"
	"path"
)

func init(){
	InitDir(CreateBlob)
}

func CreateBlob(ctx context.Context, cancel context.CancelFunc, reinit bool) error{
	dirPath := path.Join(".rec", "blob")
	return utils.CreateDir(ctx, cancel, dirPath, reinit)
}