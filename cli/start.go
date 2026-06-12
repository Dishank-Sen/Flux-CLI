package cli

import (
	"context"
	"exp1/pkg/events"
	"exp1/pkg/watcher"
	"exp1/utils/log"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func init(){
	Register("start", Start)
}

func Start() *cobra.Command{
	return &cobra.Command{
		Use: "start",
		Short: "starts recording file changes",
		RunE: startRunE,
	}
}

func startRunE(cmd *cobra.Command, args []string) error {
	// base/root context from cobra
	parentCtx := cmd.Context()

	// make a derived context that cancels on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(parentCtx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	w := watcher.NewWatcher(ctx)
	if w == nil {
		return fmt.Errorf("failed to create watcher")
	}
	ev := events.NewEvents(w, ctx)
	w.SetEvents(ev)

	// run watcher using the signal-aware ctx (not parentCtx)
	err := w.Start(ctx)

	// Always attempt to flush unsaved data
	log.Info(parentCtx, "flushing unsaved data...")
	if flushErr := ev.Flush(); flushErr != nil {
		return fmt.Errorf("flush failed: %w", flushErr)
	}

	// If Start returned an error other than cancellation, return it
	if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		return err
	}
	return nil
}
