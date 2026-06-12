package initfiles

import (
	"context"
	"exp1/cli/utils"
)

func init(){
	InitFile(CreateConfig)
}

func CreateConfig(ctx context.Context, cancel context.CancelFunc, reinit bool) error {
    return utils.CreateConfig(ctx, cancel, reinit)
}