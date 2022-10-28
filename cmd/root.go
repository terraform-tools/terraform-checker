package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const DefaultParallelism = 10

var (
	version     string
	commit      string //nolint:gochecknoglobals // don't think there's another way
	parallelism uint   //nolint:gochecknoglobals // don't think there's another way
)

func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{Use: "terraform-checker", Version: fmt.Sprintf("%s / %s", version, commit)}
	rootCmd.PersistentFlags().UintVarP(&parallelism, "parallelism", "p", DefaultParallelism, "Maximum number of terrraform parallel runs")
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := RootCmd()
	InitRootCmd(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func InitRootCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ServerCmd())
	rootCmd.AddCommand(LocalCmd())
}
