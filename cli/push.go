package cli

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	cliutils "github.com/Dishank-Sen/Flux-CLI/cli/utils"
	"github.com/Dishank-Sen/Flux-CLI/constants"
	"github.com/Dishank-Sen/Flux-CLI/types"
	"github.com/Dishank-Sen/Flux-CLI/utils"
	"github.com/Dishank-Sen/Flux-CLI/utils/logger"

	"github.com/lesismal/arpc"
	"github.com/spf13/cobra"
)

func init() {
	Register("push", Push)
}

func Push() *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "pushes all the snapshot and deltas to server",
		RunE:  pushRunE,
	}
}

func pushRunE(cmd *cobra.Command, args []string) error {
	parentCtx := cmd.Context()
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	configPath := filepath.Join(".flux", "config.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Info("no config file exist")
		logger.Info("creating default config file")

		if err := cliutils.CreateConfig(ctx, cancel, false); err != nil {
			return err
		}
	}

	cfg, err := utils.GetConfig()
	if err != nil {
		return err
	}

	// -----------------------------
	// Validate important config
	// -----------------------------
	if strings.TrimSpace(cfg.Repository.UserName) == "" {
		return fmt.Errorf("username not set. run flux set-user <username>")
	}

	if strings.TrimSpace(cfg.Repository.RemoteUrl) == "" {
		return fmt.Errorf("remote url not set. run fluxset -r <url>")
	}

	if cfg.SSHKeys.PrivateKeyPath == "" || cfg.SSHKeys.PublicKeyPath == "" {
		return fmt.Errorf("ssh keys not configured. run flux genk")
	}

	if !utils.CheckFileExist(cfg.SSHKeys.PrivateKeyPath) ||
		!utils.CheckFileExist(cfg.SSHKeys.PublicKeyPath) {
		return fmt.Errorf("ssh key files missing")
	}

	// -----------------------------
	// Authenticate user
	// -----------------------------
	logger.Info("authenticating user")

	if err := authenticateUser(cfg.Repository.UserName); err != nil {
		return err
	}

	logger.Info("authentication successful")

	// -----------------------------
	// Generate file tree
	// -----------------------------
	logger.Info("creating file tree")

	if err := cliutils.CreateFileTree(ctx); err != nil {
		return err
	}

	// -----------------------------
	// Parse remote URL
	// -----------------------------
	userName, repoName, err := parseRemoteURL(cfg.Repository.RemoteUrl)
	if err != nil {
		return err
	}

	endpointUrl := "http://localhost:3000/api/v1/push"

	// -----------------------------
	// Push snapshot
	// -----------------------------
	res, err := Trigger(userName, repoName, endpointUrl)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("response status: %s", res.Status))
	logger.Info(fmt.Sprintf("response body: %s", string(body)))

	return nil
}

func authenticateUser(username string) error {
	cli, err := arpc.NewClient(DialIPC)
	if err != nil {
		return err
	}

	req := loginRequest{
		UserName: username,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	var rsp string

	if err := cli.Call("/login", string(data), &rsp, constants.CallTime); err != nil {
		return err
	}

	var res loginResponse
	if err := json.Unmarshal([]byte(rsp), &res); err != nil {
		return err
	}

	if res.Status != 200 {
		return fmt.Errorf("authentication failed: %s", res.Message)
	}

	return nil
}

func parseRemoteURL(remoteUrl string) (username string, repoName string, err error) {
	if !strings.Contains(remoteUrl, "://") {
		remoteUrl = "https://" + remoteUrl
	}

	u, err := url.Parse(remoteUrl)
	if err != nil {
		return "", "", err
	}

	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(segments) != 2 {
		return "", "", errors.New("invalid URL format: expected /<username>/<repo>.flux")
	}

	username = segments[0]
	repo := segments[1]

	if !strings.HasSuffix(repo, ".flux") {
		return "", "", errors.New("invalid repo name: missing .flux suffix")
	}

	repoName = strings.TrimSuffix(repo, ".flux")
	if repoName == "" {
		return "", "", errors.New("empty repo name")
	}

	return username, repoName, nil
}

func Trigger(userName, repoName, endpointUrl string) (*http.Response, error) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer writer.Close()

		metaPart, _ := writer.CreateFormField("metadata")
		metadata := types.Metadata{
			UserName: userName,
			RepoName: repoName,
		}

		metadataBytes, _ := json.Marshal(metadata)
		metaPart.Write(metadataBytes)

		ignoreSet, _ := loadIgnoreSet(".flowignore")

		zipFiles(writer, "history", "history.zip", ".flux/history", ignoreSet, filterHistoryFile)
		zipFiles(writer, "fileTree", "fileTree.zip", ".flux/files", ignoreSet, nil)
		zipFiles(writer, "root-timeline", "root-timeline.zip", ".flux/root-timeline", ignoreSet, filterRootTimelineFile)
	}()

	req, _ := http.NewRequest("POST", endpointUrl, pr)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	return client.Do(req)
}

func zipFiles(
	writer *multipart.Writer,
	fieldname string,
	filename string,
	dirPath string,
	ignoreSet map[string]struct{},
	filter func([]byte, map[string]struct{}) (bool, error),
) {

	zipPart, _ := writer.CreateFormFile(fieldname, filename)
	zipWriter := zip.NewWriter(zipPart)

	filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {

		if err != nil || info.IsDir() {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if filter != nil {
			ok, err := filter(data, ignoreSet)
			if err != nil || !ok {
				return nil
			}
		}

		rel, _ := filepath.Rel(".flux", path)

		w, _ := zipWriter.Create(rel)
		w.Write(data)

		return nil
	})

	zipWriter.Close()
}

func filterHistoryFile(data []byte, ignoreSet map[string]struct{}) (bool, error) {

	var event types.Write

	if err := json.Unmarshal(data, &event); err != nil {
		return false, err
	}

	if shouldIgnore(event.Path, ignoreSet) {
		return false, nil
	}

	return true, nil
}

func filterRootTimelineFile(data []byte, ignoreSet map[string]struct{}) (bool, error) {

	var base struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(data, &base); err != nil {
		return false, err
	}

	if shouldIgnore(base.Path, ignoreSet) {
		return false, nil
	}

	return true, nil
}

func loadIgnoreSet(path string) (map[string]struct{}, error) {

	set := make(map[string]struct{})

	data, err := os.ReadFile(path)
	if err != nil {
		return set, nil
	}

	lines := strings.Split(string(data), "\n")

	for _, l := range lines {

		l = strings.TrimSpace(l)

		if l == "" {
			continue
		}

		l = filepath.Clean(l)

		set[l] = struct{}{}
	}

	return set, nil
}

func shouldIgnore(path string, ignoreSet map[string]struct{}) bool {

	path = filepath.Clean(path)

	for ignore := range ignoreSet {

		if path == ignore || strings.HasPrefix(path, ignore+string(os.PathSeparator)) {
			return true
		}
	}

	return false
}
