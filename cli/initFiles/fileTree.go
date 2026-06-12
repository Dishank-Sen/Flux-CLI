package initfiles

import (
	"context"
	"exp1/cli/utils"
)

func init(){
	InitFile(CreateFileTree)
}

// creates .rec/files/files.json which contain metadata for all files at one place
func CreateFileTree(ctx context.Context, cancel context.CancelFunc, reinit bool) error{
	return utils.CreateFileTree(ctx, cancel, reinit)
}