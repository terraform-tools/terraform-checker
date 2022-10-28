package main

import (
	"github.com/terraform-tools/terraform-checker/cmd"
	"github.com/terraform-tools/terraform-checker/pkg/logger"
)

func main() {
	logger.SetupLogger()
	cmd.Execute()
}
