package initdir

import (
	"context"
	"path"

	"github.com/Dishank-Sen/Flux-CLI/cli/utils"
)

func init() {
	InitDir(CreateFiles)
}

func CreateFiles(ctx context.Context, cancel context.CancelFunc, reinit bool) error {
	dirPath := path.Join(".flux", "files")
	return utils.CreateDir(ctx, cancel, dirPath, reinit)
}
