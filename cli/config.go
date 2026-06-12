package cli

import (
	"fmt"

	"github.com/Dishank-Sen/Flux-CLI/utils"
	"github.com/spf13/cobra"
)

func init() {
	Register("config", Config)
}

func Config() *cobra.Command {
	ConfigCmd := &cobra.Command{
		Use:   "config",
		Short: "config related operations",
		RunE:  configRunE,
	}

	ConfigCmd.Flags().BoolP("list", "l", false, "list peers")
	return ConfigCmd
}

func configRunE(cmd *cobra.Command, args []string) error {
	list, err := cmd.Flags().GetBool("list")
	if err != nil {
		return err
	}

	switch {
	case list:
		return handleList()
	default:
		return cmd.Help()
	}
}

func handleList() error {
	c, err := utils.GetConfig()
	if err != nil {
		return err
	}
	fmt.Println("Config:")
	fmt.Println(" WorkingDir:", c.WorkingDir)
	fmt.Println(" Repository:")
	fmt.Println("  UserName:", c.Repository.UserName)
	fmt.Println("  RemoteUrl:", c.Repository.RemoteUrl)
	fmt.Println(" Recorder:")
	fmt.Println("  DebounceTime:", c.Recorder.DebounceTime, "sec")
	fmt.Println("  CodeThreshold:", c.Recorder.CodeThreshold, "sec")
	return nil
}
