package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	cliUtils "github.com/Dishank-Sen/Flux-CLI/cli/utils"
	"github.com/Dishank-Sen/Flux-CLI/constants"
	"github.com/Dishank-Sen/Flux-CLI/types"
	"github.com/Dishank-Sen/Flux-CLI/utils"

	"github.com/spf13/cobra"
)

func init() {
	Register("set", Set)
}

func Set() *cobra.Command {
	SetCmd := &cobra.Command{
		Use:   "set",
		Short: "Set repository specific info",
		RunE:  setRunE,
	}

	// Define flags
	SetCmd.Flags().StringP("username", "u", "", "Git username")
	SetCmd.Flags().StringP("remoteUrl", "r", "", "Remote repository URL")
	SetCmd.Flags().String("threshold", "", "code threshold value")
	SetCmd.Flags().String("debounce", "", "debounce time")

	return SetCmd
}

func setRunE(cmd *cobra.Command, args []string) error {
	parentCtx := cmd.Context()
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()
	// Read flag values
	userName, _ := cmd.Flags().GetString("username")
	remoteUrl, _ := cmd.Flags().GetString("remoteUrl")
	threshold, _ := cmd.Flags().GetString("threshold")
	debounce, _ := cmd.Flags().GetString("debounce")

	configPath := filepath.Join(".flux", "config.json")

	// check if config.json exists
	if !utils.CheckFileExist(configPath) {
		// create a config.json with empty entries
		err := cliUtils.CreateConfig(ctx, cancel, false)
		if err != nil {
			return err
		}
	}

	// Load existing config if present
	var config types.Config
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			if uerr := json.Unmarshal(data, &config); uerr != nil {
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
	if strings.TrimSpace(threshold) != "" {
		threshold_i, err := strconv.Atoi(threshold)
		if err != nil {
			return err
		}
		if verifyThreshold(int16(threshold_i)) {
			config.Recorder.CodeThreshold = int16(threshold_i)
		} else {
			return fmt.Errorf("this threshold is not acceptable")
		}
	}
	if strings.TrimSpace(debounce) != "" {
		debounce_i, err := strconv.Atoi(debounce)
		if err != nil {
			return err
		}
		if verifyDebounce(int16(debounce_i)) {
			config.Recorder.CodeThreshold = int16(debounce_i)
		} else {
			return fmt.Errorf("this debounce time is not acceptable")
		}
	}

	// Write back to config.json
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error (set.go): %s", err.Error())
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error (set.go): %s", err.Error())
	}

	fmt.Println("✅ Repository configuration updated successfully!")

	return nil
}

func verifyThreshold(threshold int16) bool {
	if threshold <= 0 || threshold > constants.MaxThreshold {
		return false
	}
	return true
}

func verifyDebounce(debounce int16) bool {
	if debounce <= 0 || debounce > constants.MaxDebounce {
		return false
	}
	return true
}
