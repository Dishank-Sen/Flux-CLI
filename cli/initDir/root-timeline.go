package initdir

import (
	"context"
	"exp1/cli/utils"
	"path"
)

func init(){
	InitDir(CreateRootTimeline)
}

func CreateRootTimeline(ctx context.Context, cancel context.CancelFunc, reinit bool) error{
	dirPath := path.Join(".rec", "root-timeline")
	return utils.CreateDir(ctx, cancel, dirPath, reinit)
}