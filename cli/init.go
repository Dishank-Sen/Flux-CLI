package cli

import (
	"context"
	"errors"
	initdir "exp1/cli/initDir"
	initfiles "exp1/cli/initFiles"
	"exp1/utils"
	"exp1/utils/log"

	"github.com/spf13/cobra"
)

var ErrSkipRun = errors.New("cli: skip runE")

func init(){
	Register("init", Init)
}

func Init() *cobra.Command{
	return &cobra.Command{
		Use: "init",
		Short: "initialize a new rec repository",
		RunE: initRunE,
		PersistentPreRunE: initPersistentPreRunE,
		SilenceUsage: true,     // prevents usage on error
		SilenceErrors: true,    // prevents printing sentinel error
	}
}

func initPersistentPreRunE(cmd *cobra.Command, args []string)error{
	rootDir := ".rec"
	parentCtx := cmd.Context()
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	if utils.CheckDirExist(rootDir){
		log.Info(parentCtx, "Reinitializing rec repository")
		if err := reinitialize(ctx, cancel); err != nil {
			return err // real error
		}
		return ErrSkipRun // signals to skip RunE
	}

	return nil
}

func initRunE(cmd *cobra.Command, args []string) error{
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

    // Create directories
    err := createDir(ctx, cancel, false)
	if err != nil{
		return err
	}

    // Create files
    err = createFiles(ctx, cancel, false)
	if err != nil{
		return err
	}

	log.Info(ctx, "Initialized empty rec repository")
	return nil
}

func createFiles(ctx context.Context, cancel context.CancelFunc, reinit bool) error{
	for _, f := range initfiles.InitFiles{
		err := f(ctx, cancel, reinit)
		if err != nil{
			return err
		}
	}
	return nil
}

func createDir(ctx context.Context, cancel context.CancelFunc, reinit bool) error{
	for _, f := range initdir.InitDirectories{
		err := f(ctx, cancel, reinit)
		if err != nil{
			return err
		}
	}
	return nil
}

func reinitialize(ctx context.Context, cancel context.CancelFunc) error{
    // Create directories
    err := createDir(ctx, cancel, true)
	if err != nil{
		return err
	}

    // Create files
    err = createFiles(ctx, cancel, true)
	if err != nil{
		return err
	}

	log.Info(ctx, "Reinitialized rec repository")
	return nil
}