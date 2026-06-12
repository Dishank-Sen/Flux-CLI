package constants

import (
	"os"
	"path"
	"path/filepath"
	"time"
)

const (
	MaxThreshold = 10 * 1000
	MaxDebounce  = 1 * 60 // 1 min
	CallTime     = time.Second * 10
)

func PrivateKeyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "share", "flux-daemon", "ssh-keys", "id_ed25519"), nil
}

func PublicKeyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "share", "flux-daemon", "ssh-keys", "id_ed25519.pub"), nil
}

func SSHKeysDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(home, ".local", "share", "flux-daemon", "ssh-keys"), nil
}
