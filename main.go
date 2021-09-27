package main

import (
	"github.com/terraform-tools/terraform-checker/pkg/logger"
	"github.com/terraform-tools/terraform-checker/pkg/server"
)

func main() {
	logger.SetupLogger()
	server.StartServer()
}
