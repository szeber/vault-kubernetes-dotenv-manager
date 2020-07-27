package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/szeber/vault-kubernetes-dotenv-manager/config"
	"github.com/szeber/vault-kubernetes-dotenv-manager/constants"
	"github.com/szeber/vault-kubernetes-dotenv-manager/helper"
	"github.com/szeber/vault-kubernetes-dotenv-manager/secret_manager"
	"os"
)

var configPath = flag.String("config", "config.yaml", "The path to the config file")
var mode = flag.String("mode", "", "The operating mode. Required. Valid values are 'populate' or 'keep-alive'")
var httpPort = flag.Int("http-port", 8000, "The HTTP port for liveness and readiness checks")

func main() {
	flag.Parse()

	if !helper.StringInSlice(constants.ValidModes[:], *mode) {
		flag.Usage()
		glog.Exit("Invalid mode or no mode set")
	}

	if *flag.Bool("help", false, "Show help") {
		flag.Usage()
		os.Exit(0)
	}

	appConfig := config.LoadConfig(*configPath)

	switch *mode {
	case constants.ModePopulate:
		secret_manager.PopulateSecrets(appConfig)
	case constants.ModeKeepAlive:
		secret_manager.KeepSecretsAlive(appConfig, *httpPort)
	default:
		glog.Exit("Invalid operating mode: " + *mode)
	}
}
