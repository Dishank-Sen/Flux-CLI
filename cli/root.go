package cli

import (
	"context"
	"exp1/utils"
	"fmt"

	"github.com/spf13/cobra"
)

func Root(ctx context.Context) *cobra.Command{
	var rootCmd = &cobra.Command{
		Use: "rec",
		Short: "rec is a simple version control system",
		Long: "rec is a version control system built in Go. It captures code changes.",
		PersistentPreRunE: persistentPreRunE,
	}

	rootCmd.SetContext(ctx)

	// loop which register all the commands
	for _, cmd := range Registered{
		c := cmd()
		rootCmd.AddCommand(c)
	}

	return rootCmd
}

func persistentPreRunE(cmd *cobra.Command, args []string) error{
	// msg := fmt.Sprintf("cmd name: %s", cmd.Name())
	// log.Info(cmd.Context(), msg)
	if cmd.Name() == "init"{
		return nil
	}
	
	// if .rec is not created prompt user to run init command
	if !utils.CheckDirExist(".rec"){
		// log.Info(cmd.Context(), "debug-2")
		err := "not a rec repository, run 'rec init' to initialize a empty rec repository"
		return fmt.Errorf(err)
	}
	
	return nil
}