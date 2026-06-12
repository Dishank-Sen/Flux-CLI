package cli

import (
	"archive/zip"
	"context"
	"encoding/json"
	"exp1/cli/utils"
	"exp1/internal/types"
	"exp1/utils/log"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init(){
	Register("push", Push)
}

func Push() *cobra.Command{
	return &cobra.Command{
		Use: "push",
		Short: "pushes all the snapshot and deltas to server",
		RunE: pushRunE,
	}
}

func pushRunE(cmd *cobra.Command, args []string) error{
	configPath := filepath.Join(".rec", "config.json")
	parentCtx := cmd.Context()
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("No .rec/config.json found. Run 'rec init' and 'rec set --remoteUrl <url>' first.")
		log.Info(parentCtx, "no config file exist")
		log.Info(parentCtx, "creating default config file.")

		// create a default config file
		err := utils.CreateConfig(ctx, cancel, false)
		if err != nil{
			return err
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config types.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
 
	remoteUrl := config.Repository.RemoteUrl

	if strings.TrimSpace(remoteUrl) == "" {
		return fmt.Errorf("no remote url found, run rec set -r <remoteUrl> to set it.")
	}

	res, err := Trigger(remoteUrl)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	fmt.Println("response status:", res.Status)
	fmt.Println("response body:", string(body))
	statusMsg := fmt.Sprintf("response status: %s", res.Status)
	log.Info(parentCtx, statusMsg)

	bodyMsg := fmt.Sprintf("response body: %s", string(body))
	log.Info(parentCtx, bodyMsg)

	return nil
}

func Trigger(remoteUrl string) (*http.Response, error){
	// get pipe reader and writer
	pr, pw := io.Pipe()

	// get a writer to the pipe
	zipWriter := zip.NewWriter(pw)

	go func(){
		filepath.Walk(".rec/history", func(path string, info fs.FileInfo, err error) error {
			if info.IsDir(){
				return nil
			}

			f, err := os.Open(path)
			if err != nil{
				return err
			}

			rel, err := filepath.Rel(".rec", path)
			if err != nil{
				return err
			}

			w, err := zipWriter.Create(rel)
			
			io.Copy(w, f)
			return nil
		})
		zipWriter.Close()
		pw.Close()
	}()

	return http.Post(remoteUrl, "application/zip", pr)
}