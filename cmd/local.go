package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/terraform-tools/terraform-checker/pkg/local"
)

func LocalCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "local",
		Short: "run terraform-checker in local mode",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("you must provide a path as first and only arg")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			local.StartLocal(args[0], parallelism)
		},
	}
}
