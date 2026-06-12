package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	initdir "github.com/Dishank-Sen/Flux-CLI/cli/initDir"
	initfiles "github.com/Dishank-Sen/Flux-CLI/cli/initFiles"
	"github.com/Dishank-Sen/Flux-CLI/constants"
	"github.com/Dishank-Sen/Flux-CLI/utils"
	"github.com/Dishank-Sen/Flux-CLI/utils/logger"

	"github.com/lesismal/arpc"
	"github.com/spf13/cobra"
)

type initRequest struct {
	WorkingDir  string
	IgnoreFiles []string
}

type initResponse struct {
	Status  int16
	Message string
}

var ErrSkipRun = errors.New("cli: skip runE")

func init() {
	Register("init", Init)
}

func Init() *cobra.Command {
	return &cobra.Command{
		Use:               "init",
		Short:             "initialize a new flux repository",
		RunE:              initRunE,
		PersistentPreRunE: initPersistentPreRunE,
		SilenceUsage:      true, // prevents usage on error
		SilenceErrors:     true, // prevents printing sentinel error
	}
}

func initPersistentPreRunE(cmd *cobra.Command, args []string) error {
	rootDir := ".flux"
	parentCtx := cmd.Context()
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	if utils.CheckDirExist(rootDir) {
		logger.Info("Reinitializing flux repository")
		if err := reinitialize(ctx, cancel); err != nil {
			return err // real error
		}
		return ErrSkipRun // signals to skip RunE
	}

	return nil
}

func initRunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	cli, err := arpc.NewClient(DialIPC)
	if err != nil {
		return err
	}

	req, err := getInitReq()
	if err != nil {
		return err
	}
	rsp := ""
	if err := cli.Call("/init", &req, &rsp, constants.CallTime); err != nil {
		return err
	} else {
		var res initResponse
		if err := json.Unmarshal([]byte(rsp), &res); err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("Status: %v\nMessage: %s", res.Status, res.Message))

		// Create directories
		err := createDir(ctx, cancel, false)
		if err != nil {
			return err
		}

		// Create files
		if err := createFiles(ctx, cancel, false); err != nil {
			return err
		}
	}
	logger.Info("Initialized empty flux repository")
	return nil
}

func createFiles(ctx context.Context, cancel context.CancelFunc, reinit bool) error {
	for _, f := range initfiles.InitFiles {
		err := f(ctx, cancel, reinit)
		if err != nil {
			return err
		}
	}
	return nil
}

func createDir(ctx context.Context, cancel context.CancelFunc, reinit bool) error {
	for _, f := range initdir.InitDirectories {
		err := f(ctx, cancel, reinit)
		if err != nil {
			return err
		}
	}
	return nil
}

func reinitialize(ctx context.Context, cancel context.CancelFunc) error {
	// Create directories
	err := createDir(ctx, cancel, true)
	if err != nil {
		return err
	}

	// Create files
	err = createFiles(ctx, cancel, true)
	if err != nil {
		return err
	}

	logger.Info("Reinitialized flux repository")
	return nil
}

func getInitReq() (string, error) {
	f := utils.GetIgnoreFiles()
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	req := initRequest{
		WorkingDir:  dir,
		IgnoreFiles: f,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
