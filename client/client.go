package client

import (
	"net"
	"runtime"

	"github.com/lesismal/arpc"
)

type Client struct {
	client *arpc.Client
}

func NewClient() (*Client, error) {
	c, err := arpc.NewClient(dialIPC)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: c,
	}, nil
}

func dialIPC() (net.Conn, error) {
	switch runtime.GOOS {

	case "windows":
		// local TCP
		return net.Dial("tcp", "127.0.0.1:43899")

	case "linux", "darwin":
		return net.Dial("unix", "/tmp/flux.sock")

	default:
		return net.Dial("tcp", "127.0.0.1:43899")
	}
}
