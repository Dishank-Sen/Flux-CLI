package initdir

import (
	"context"
	"path"

	"github.com/Dishank-Sen/Flux-CLI/cli/utils"
)

func init() {
	InitDir(CreateRootTimeline)
}

func CreateRootTimeline(ctx context.Context, cancel context.CancelFunc, reinit bool) error {
	dirPath := path.Join(".flux", "root-timeline")
	return utils.CreateDir(ctx, cancel, dirPath, reinit)
}
