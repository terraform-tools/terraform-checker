package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/terraform-tools/terraform-checker/pkg/filter"
	"github.com/terraform-tools/terraform-checker/pkg/local"
	"github.com/terraform-tools/terraform-checker/pkg/terraform"
)

var (
	fmtCheck      bool //nolint:gochecknoglobals // don't think there's another way
	validateCheck bool //nolint:gochecknoglobals // don't think there's another way
	tfLintCheck   bool //nolint:gochecknoglobals // don't think there's another way
)

func LocalCmd() *cobra.Command {
	localCmd := &cobra.Command{
		Use:   "local",
		Short: "run terraform-checker in local mode",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("you must provide a path as first and only arg")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			tfChecksTypes := []string{}
			if fmtCheck {
				tfChecksTypes = append(tfChecksTypes, terraform.Fmt.String())
			}
			if validateCheck {
				tfChecksTypes = append(tfChecksTypes, terraform.Validate.String())
			}
			if tfLintCheck {
				tfChecksTypes = append(tfChecksTypes, terraform.TFLint.String())
			}

			if len(tfChecksTypes) == 0 {
				tfChecksTypes = terraform.AllTfCheckTypes()
			}

			local.StartLocal(args[0], parallelism, filter.TfCheckTypeFilter{TfCheckTypes: tfChecksTypes})
		},
	}
	localCmd.PersistentFlags().BoolVarP(&fmtCheck, "fmt", "", false, "Whether to execute fmt check or not")
	localCmd.PersistentFlags().BoolVarP(&validateCheck, "validate", "", false, "Whether to execute validate check or not")
	localCmd.PersistentFlags().BoolVarP(&tfLintCheck, "tflint", "", false, "Whether to execute tflint check or not")
	return localCmd
}
