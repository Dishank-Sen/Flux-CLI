package utils

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/Dishank-Sen/Flux-CLI/types"
	"github.com/Dishank-Sen/Flux-CLI/utils/logger"
)

func CheckDirExist(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	if info.IsDir() {
		return true
	}
	return false
}

func CheckFileExist(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func CreateFile(ctx context.Context, cancel context.CancelFunc, path string) {
	f, err := os.Create(path)
	if err != nil {
		logger.Error(err.Error())
	}
	f.Close()
}

func GetConfig() (types.Config, error) {
	dir, err := os.Getwd()
	if err != nil {
		return types.Config{}, err
	}
	filePath := path.Join(dir, ".flux", "config.json")
	byteData, err := os.ReadFile(filePath)
	if err != nil {
		return types.Config{}, err
	}
	var cfg types.Config
	if err := json.Unmarshal(byteData, &cfg); err != nil {
		return types.Config{}, err
	}
	return cfg, nil
}

func GetIgnoreFiles() []string {
	byteData, err := os.ReadFile(".flowignore")
	if err != nil {
		return []string{".flux"}
	}
	data := string(byteData)
	parts := strings.Split(data, "\n")
	if !slices.Contains(parts, ".flux") {
		parts = append(parts, ".flux")
	}
	return parts
}

func OverwriteFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}
