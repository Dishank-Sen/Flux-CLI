package recorder

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func HashFile(path string) (string, error) {
    file, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer file.Close()

    // Check file size
    stat, err := file.Stat()
    if err != nil {
        return "", err
    }
    if stat.Size() == 0 {
        // empty file => return empty string
        return "", nil 
    }

    // Hash non-empty file
    hasher := sha256.New()
    if _, err := io.Copy(hasher, file); err != nil {
        return "", err
    }

    hash := hasher.Sum(nil)
    return hex.EncodeToString(hash), nil
}