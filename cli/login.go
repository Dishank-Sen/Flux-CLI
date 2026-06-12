package cli

import (
	"encoding/json"
	"fmt"

	"github.com/Dishank-Sen/Flux-CLI/constants"
	"github.com/Dishank-Sen/Flux-CLI/utils"
	"github.com/Dishank-Sen/Flux-CLI/utils/logger"

	"github.com/lesismal/arpc"
	"github.com/spf13/cobra"
)

type loginResponse struct {
	Status  int16
	Message string
}

type loginRequest struct {
	UserName string
}

func init() {
	Register("login", Login)
}

func Login() *cobra.Command {
	LoginCmd := &cobra.Command{
		Use:   "login",
		Short: "authenticates user",
		RunE:  loginRunE,
	}

	return LoginCmd
}

func loginRunE(cmd *cobra.Command, args []string) error {
	cli, err := arpc.NewClient(DialIPC)
	if err != nil {
		return err
	}

	req, err := getLoginReq()
	if err != nil {
		return err
	}
	rsp := ""
	err = cli.Call("/login", &req, &rsp, constants.CallTime)
	if err != nil {
		return err
	} else {
		var res loginResponse
		if err := json.Unmarshal([]byte(rsp), &res); err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("Status: %v\nMessage: %s", res.Status, res.Message))
	}
	return nil
}

func getLoginReq() (string, error) {
	cfg, err := utils.GetConfig()
	req := loginRequest{
		UserName: cfg.Repository.UserName,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
