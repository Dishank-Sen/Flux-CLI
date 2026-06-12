package cli

import (
	"context"
	"exp1/utils"
	"fmt"

	"github.com/spf13/cobra"
)

func Root(ctx context.Context) *cobra.Command{
	var rootCmd = &cobra.Command{
		Use: "flux",
		Short: "flux is a simple version control system",
		Long: "flux is a version control system built in Go. It captures code changes.",
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
	
	// if .flux is not created prompt user to run init command
	if !utils.CheckDirExist(".flux"){
		// log.Info(cmd.Context(), "debug-2")
		err := "not a flux repository, run 'flux init' to initialize a empty flux repository"
		return fmt.Errorf("Error: %s", err)
	}
	
	return nil
}