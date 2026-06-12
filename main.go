package main

import (
	"context"
	"errors"
	"exp1/cli"
	"exp1/utils/log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)


func main(){
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := godotenv.Load()
	if err != nil {
		log.Error(ctx, stop, err.Error())	
	}

	rootCmd := cli.Root(ctx)
	if err := rootCmd.Execute(); err != nil {
    if errors.Is(err, cli.ErrSkipRun) {
        // normal reinit completed, exit 0 (no message)
        os.Exit(0)
    }
    log.Error(ctx, stop, err.Error())
    os.Exit(1)
}
}