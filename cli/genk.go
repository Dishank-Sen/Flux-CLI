package cli

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cliUtils "github.com/Dishank-Sen/Flux-CLI/cli/utils"
	"github.com/Dishank-Sen/Flux-CLI/constants"
	"github.com/Dishank-Sen/Flux-CLI/utils"
	"github.com/Dishank-Sen/Flux-CLI/utils/logger"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func init() {
	Register("genk", Genk)
}

func Genk() *cobra.Command {
	GenkCmd := &cobra.Command{
		Use:   "genk",
		Short: "generate private and public key",
		RunE:  genkRunE,
	}

	GenkCmd.Flags().StringP("sync", "s", "", "sync config with ssh keys already at its path")

	return GenkCmd
}

func genkRunE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	syncPath, _ := cmd.Flags().GetString("sync")

	cfgPath := filepath.Join(".flux", "config.json")
	if !utils.CheckFileExist(cfgPath) {
		logger.Info("config not exist creating config...")
		if err := cliUtils.CreateConfig(ctx, cancel, true); err != nil {
			return err
		}
	}

	cfg, err := utils.GetConfig()
	if err != nil {
		return err
	}

	dir, err := constants.SSHKeysDir()
	if err != nil {
		return err
	}

	// ---- SYNC EXISTING KEYS ----
	if syncPath != "" {
		if err := syncExistingKeys(syncPath, dir); err != nil {
			return err
		}

		if err := updateConfig(dir); err != nil {
			return err
		}

		logger.Info("ssh keys synced successfully")
		return nil
	}

	// ---- GENERATE NEW KEYS ----
	if cfg.SSHKeys.PrivateKeyPath != "" && cfg.SSHKeys.PublicKeyPath != "" {
		if utils.CheckFileExist(cfg.SSHKeys.PrivateKeyPath) && utils.CheckFileExist(cfg.SSHKeys.PublicKeyPath) {
			logger.Info("keys already exists")
			return nil
		}
	}

	email, err := promptEmail()
	if err != nil {
		return err
	}

	if err := generateSSHKeyPair(dir, email); err != nil {
		return err
	}

	if err := updateConfig(dir); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("generated ssh key pair in directory: %s", dir))

	return nil
}

func syncExistingKeys(srcDir string, dstDir string) error {
	privateSrc := filepath.Join(srcDir, "id_ed25519")
	publicSrc := filepath.Join(srcDir, "id_ed25519.pub")

	if !utils.CheckFileExist(privateSrc) || !utils.CheckFileExist(publicSrc) {
		return errors.New("id_ed25519 or id_ed25519.pub not found in provided directory")
	}

	// validate private key
	privData, err := os.ReadFile(privateSrc)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(privData)
	if block == nil || block.Type != "OPENSSH PRIVATE KEY" {
		return errors.New("invalid OpenSSH private key format")
	}

	// validate public key
	pubData, err := os.ReadFile(publicSrc)
	if err != nil {
		return err
	}

	if _, _, _, _, err := ssh.ParseAuthorizedKey(pubData); err != nil {
		return errors.New("invalid SSH public key format")
	}

	// ensure destination dir exists
	if err := os.MkdirAll(dstDir, 0700); err != nil {
		return err
	}

	privateDst := filepath.Join(dstDir, "id_ed25519")
	publicDst := filepath.Join(dstDir, "id_ed25519.pub")

	if err := copyFile(privateSrc, privateDst, 0600); err != nil {
		return err
	}

	if err := copyFile(publicSrc, publicDst, 0644); err != nil {
		return err
	}

	return nil
}

func copyFile(src string, dst string, perm os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, perm)
}

func generateSSHKeyPair(dir string, email string) error {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return err
	}

	pubBytes := ssh.MarshalAuthorizedKey(sshPub)
	pubKey := fmt.Sprintf("%s %s\n", pubBytes[:len(pubBytes)-1], email)

	privBytes, err := ssh.MarshalPrivateKey(priv, email)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	privPath := filepath.Join(dir, "id_ed25519")
	if err := os.WriteFile(privPath, pem.EncodeToMemory(privBytes), 0600); err != nil {
		return err
	}

	pubPath := filepath.Join(dir, "id_ed25519.pub")
	if err := os.WriteFile(pubPath, []byte(pubKey), 0644); err != nil {
		return err
	}

	return nil
}

func promptEmail() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(email), nil
}

func updateConfig(dir string) error {
	privateKeyPath := filepath.Join(dir, "id_ed25519")
	publicKeyPath := filepath.Join(dir, "id_ed25519.pub")
	cfgPath := filepath.Join(".flux", "config.json")

	cfg, err := utils.GetConfig()
	if err != nil {
		return err
	}

	cfg.SSHKeys.PrivateKeyPath = privateKeyPath
	cfg.SSHKeys.PublicKeyPath = publicKeyPath

	newCfg, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		return err
	}

	return utils.OverwriteFile(cfgPath, newCfg)
}
