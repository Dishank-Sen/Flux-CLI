package initdir

import (
	"context"
	"exp1/cli/utils"
	"path"
)

func init(){
	InitDir(CreateHistory)
}

func CreateHistory(ctx context.Context, cancel context.CancelFunc, reinit bool) error{
	dirPath := path.Join(".rec", "history")
	return utils.CreateDir(ctx, cancel, dirPath, reinit)
}