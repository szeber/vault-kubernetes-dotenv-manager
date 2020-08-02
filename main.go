package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/szeber/vault-kubernetes-dotenv-manager/config"
	"github.com/szeber/vault-kubernetes-dotenv-manager/constants"
	"github.com/szeber/vault-kubernetes-dotenv-manager/helper"
	"github.com/szeber/vault-kubernetes-dotenv-manager/secret_manager"
	"os"
	"os/signal"
	"syscall"
)

var configPath = flag.String("config", "config.yaml", "The path to the config file")
var mode = flag.String("mode", "", "The operating mode. Optional. Valid values are 'populate' or 'keep-alive'. Defaults to doing both operations")
var httpPort = flag.Int("http-port", 8000, "The HTTP port for liveness and readiness checks")

func main() {
	// Set default values
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "Info")
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
	case "":
		secret_manager.PopulateSecrets(appConfig)
		revokeAuthLeaseOnQuit(appConfig)
		secret_manager.KeepSecretsAlive(appConfig, *httpPort)
	case constants.ModePopulate:
		secret_manager.PopulateSecrets(appConfig)
	case constants.ModeKeepAlive:
		revokeAuthLeaseOnQuit(appConfig)
		secret_manager.KeepSecretsAlive(appConfig, *httpPort)
	default:
		glog.Exit("Invalid operating mode: " + *mode)
	}
}

func revokeAuthLeaseOnQuit(appConfig config.Config) {
	if !appConfig.RevokeAuthLeaseOnQuit {
		return
	}

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		glog.Info("Signal received, shutting down: ", sig)
		glog.Info("Revoking auth lease")
		secret_manager.RevokeAuthLease(appConfig)
		os.Exit(0)
	}()
}
