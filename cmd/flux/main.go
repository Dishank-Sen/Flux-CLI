package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/Dishank-Sen/Flux-CLI/cli"
	"github.com/Dishank-Sen/Flux-CLI/utils/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rootCmd := cli.Root(ctx)
	if err := rootCmd.Execute(); err != nil {
		if errors.Is(err, cli.ErrSkipRun) {
			// normal reinit completed, exit 0 (no message)
			os.Exit(0)
		}
		logger.Error(err.Error())
		stop()
	}
}
