package cli

import (
	"fmt"
	"os"

	"github.com/Dishank-Sen/Flux-CLI/utils"
	"github.com/Dishank-Sen/Flux-CLI/utils/logger"

	"github.com/spf13/cobra"
)

func init() {
	Register("showk", Showk)
}

func Showk() *cobra.Command {
	ShowkCmd := &cobra.Command{
		Use:   "showk",
		Short: "display ssh public and private keys",
		RunE:  showkRunE,
	}

	return ShowkCmd
}

func showkRunE(cmd *cobra.Command, args []string) error {
	cfg, err := utils.GetConfig()
	if err != nil {
		return err
	}

	if cfg.SSHKeys.PrivateKeyPath == "" || cfg.SSHKeys.PublicKeyPath == "" {
		return fmt.Errorf("ssh keys not configured")
	}

	privKey, err := os.ReadFile(cfg.SSHKeys.PrivateKeyPath)
	if err != nil {
		return err
	}

	pubKey, err := os.ReadFile(cfg.SSHKeys.PublicKeyPath)
	if err != nil {
		return err
	}

	logger.Info("Private Key:")
	fmt.Println(string(privKey))

	logger.Info("Public Key:")
	fmt.Println(string(pubKey))

	return nil
}
