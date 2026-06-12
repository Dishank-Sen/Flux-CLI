package cli

import (
	"github.com/spf13/cobra"
)

type cmdFunc func() *cobra.Command
var Registered map[string]cmdFunc

func Register(cmd string, f cmdFunc){
	if Registered == nil{
		Registered = make(map[string]cmdFunc)
	}
	Registered[cmd] = f
}