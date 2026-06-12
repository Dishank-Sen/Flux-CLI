package cli

import (
	"context"
	"fmt"

	"github.com/Dishank-Sen/Flux-CLI/utils"

	"github.com/spf13/cobra"
)

func Root(ctx context.Context) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:               "flux",
		Short:             "flux is a simple version control system",
		Long:              "flux is a version control system built in Go. It captures code changes.",
		PersistentPreRunE: persistentPreRunE,
	}

	rootCmd.SetContext(ctx)

	// loop which register all the commands
	for _, cmd := range Registered {
		c := cmd()
		rootCmd.AddCommand(c)
	}

	return rootCmd
}

func persistentPreRunE(cmd *cobra.Command, args []string) error {
	if cmd.Name() == "init" {
		return nil
	}

	// if .fluxis not created prompt user to run init command
	if !utils.CheckDirExist(".flux") {
		err := "not a flux repository, run 'flux init' to initialize a empty flux repository"
		return fmt.Errorf("%s", err)
	}

	return nil
}
