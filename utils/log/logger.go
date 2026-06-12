package log

import (
	"context"
	"fmt"
	"os"
)

// Error cancels the context (triggering cleanup), prints an error, and exits
func Error(ctx context.Context, cancel context.CancelFunc, msg string) {
	fmt.Println("Error[CLI]:", msg)

	// trigger cancellation so all goroutines watching ctx.Done() exit safely
	if cancel != nil {
		cancel()
	}

	// let defers run in the caller, then exit
	os.Exit(1)
}

// Info prints informational messages
func Info(ctx context.Context, msg string) {
	// optional: skip logging if shutting down
	if ctx.Err() != nil {
		return
	}

	fmt.Println("Info[CLI]:", msg)
}
