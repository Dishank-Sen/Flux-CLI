package cli

import (
	"context"
	"encoding/json"
	"exp1/internal/types"
	"exp1/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	cliUtils "exp1/cli/utils"
	"github.com/spf13/cobra"
)

func init(){
	Register("set", Set)
}

func Set() *cobra.Command{
	SetCmd := &cobra.Command{
		Use:   "set",
		Short: "Set repository specific info",
		RunE:   setRunE,
	}

	// Define flags
	SetCmd.Flags().StringP("username", "u", "", "username")
	SetCmd.Flags().StringP("remoteUrl", "r", "", "Remote repository URL")

	return SetCmd
}

func setRunE(cmd *cobra.Command, args []string) error{
	parentCtx := cmd.Context()
	ctx, cancel := 	context.WithCancel(parentCtx)
	defer cancel()
	// Read flag values
	userName, _ := cmd.Flags().GetString("username")
	remoteUrl, _ := cmd.Flags().GetString("remoteUrl")

	configPath := filepath.Join(".rec", "config.json")

	// check if config.json exists
	if !utils.CheckFileExist(configPath){
		// create a config.json with empty entries
		err := cliUtils.CreateConfig(ctx, cancel, false)
		if err != nil{
			return err
		}
	}

	// Load existing config if present
	var config types.Config
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			if uerr := json.Unmarshal(data, &config); uerr != nil{
				return uerr
			}
		}
	}

	// Update only provided fields
	if strings.TrimSpace(userName) != "" {
		config.Repository.UserName = userName
	}
	if strings.TrimSpace(remoteUrl) != "" {
		config.Repository.RemoteUrl = remoteUrl
	}

	// Write back to config.json
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error (set.go): %s",err.Error())
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error (set.go): %s",err.Error())
	}

	fmt.Println("✅ Repository configuration updated successfully!")

	return nil
}
