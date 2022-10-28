package cmd

import (
	"github.com/spf13/cobra"
	"github.com/terraform-tools/terraform-checker/pkg/server"
)

func ServerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "run terraform-checker in server mode",
		Run: func(cmd *cobra.Command, args []string) {
			server.StartServer()
		},
	}
}
