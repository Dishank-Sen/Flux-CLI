package initdir

import (
	"context"
	"path"

	"github.com/Dishank-Sen/Flux-CLI/cli/utils"
)

func init() {
	InitDir(CreateHistory)
}

func CreateHistory(ctx context.Context, cancel context.CancelFunc, reinit bool) error {
	dirPath := path.Join(".flux", "history")
	return utils.CreateDir(ctx, cancel, dirPath, reinit)
}
