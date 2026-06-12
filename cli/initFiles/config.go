package initfiles

import (
	"context"

	"github.com/Dishank-Sen/Flux-CLI/cli/utils"
)

func init() {
	InitFile(CreateConfig)
}

func CreateConfig(ctx context.Context, cancel context.CancelFunc, reinit bool) error {
	return utils.CreateConfig(ctx, cancel, reinit)
}
