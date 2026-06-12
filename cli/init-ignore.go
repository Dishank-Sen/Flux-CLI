package cli

import (
	"fmt"
	"os"

	"github.com/Dishank-Sen/Flux-CLI/utils/logger"

	"github.com/spf13/cobra"
)

func init() {
	Register("init-ignore", InitIgnore)
}

func InitIgnore() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-ignore",
		Short: "generate a default .flowignore file",
		RunE:  initIgnoreRunE,
	}

	return cmd
}

func initIgnoreRunE(cmd *cobra.Command, args []string) error {
	ignoreFile := ".flowignore"

	// don't overwrite if file already exists
	if _, err := os.Stat(ignoreFile); err == nil {
		logger.Info(".flowignore already exists")
		return nil
	}

	defaultContent := `# flux ignore rules

# build artifacts
bin
build
dist

# dependency directories
vendor

# VCS
.git
.gitignore

# flux internal
.flux

# logs
*.log

# temp files
*.tmp
*.swp

# OS files
.DS_Store
Thumbs.db
`

	err := os.WriteFile(ignoreFile, []byte(defaultContent), 0644)
	if err != nil {
		return err
	}

	logger.Info("created default .flowignore file")
	fmt.Println(".flowignore generated successfully")

	return nil
}
