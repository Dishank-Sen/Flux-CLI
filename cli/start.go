package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Dishank-Sen/Flux-CLI/constants"
	"github.com/Dishank-Sen/Flux-CLI/utils"
	"github.com/Dishank-Sen/Flux-CLI/utils/logger"
	"github.com/lesismal/arpc"
	"github.com/spf13/cobra"
)

type startRequest struct {
	WorkingDir  string
	IgnoreFiles []string
}

type startResponse struct {
	Status  int16
	Message string
}

func init() {
	Register("start", Start)
}

func Start() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "starts recording file changes",
		RunE:  startRunE,
	}
}

func startRunE(cmd *cobra.Command, args []string) error {
	cli, err := arpc.NewClient(DialIPC)
	if err != nil {
		return err
	}

	req, err := getStartReq()
	if err != nil {
		return err
	}
	rsp := ""
	err = cli.Call("/start", &req, &rsp, constants.CallTime)
	if err != nil {
		return err
	} else {
		var res startResponse
		if err := json.Unmarshal([]byte(rsp), &res); err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("Status: %v\nMessage: %s", res.Status, res.Message))
	}

	// w := watcher.NewWatcher(ctx)
	// if w == nil {
	// 	return fmt.Errorf("failed to create watcher")
	// }
	// ev := events.NewEvents(w, ctx)
	// w.SetEvents(ev)

	// // run watcher using the signal-aware ctx (not parentCtx)
	// err := w.Start(ctx)

	// // Always attempt to flush unsaved data
	// log.Info(parentCtx, "flushing unsaved data...")
	// if flushErr := ev.Flush(); flushErr != nil {
	// 	return fmt.Errorf("flush failed: %w", flushErr)
	// }

	// // If Start returned an error other than cancellation, return it
	// if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
	// 	return err
	// }
	return nil
}

func getStartReq() (string, error) {
	f := utils.GetIgnoreFiles()
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	req := startRequest{
		WorkingDir:  dir,
		IgnoreFiles: f,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
