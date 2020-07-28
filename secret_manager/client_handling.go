package secret_manager

import (
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	"github.com/szeber/vault-kubernetes-dotenv-manager/config"
	"github.com/szeber/vault-kubernetes-dotenv-manager/data"
	"github.com/szeber/vault-kubernetes-dotenv-manager/vault"
	"time"
)

var apiClientAuthLifetimeSeconds int
var apiClientAuthLifetime time.Time
var apiClient *api.Client

func getClient(appConfig config.Config, existingToken string) *api.Client {
	if isCurrentApiClientValid() {
		return apiClient
	}

	if "" == existingToken {
		return makeClient(appConfig)
	}

	var err error
	apiClient, err = vault.GetClientWithToken(appConfig, existingToken)

	if err != nil {
		glog.Exit("Failed to get vault client with existing token: ", err)
	}

	return apiClient
}

func makeClient(appConfig config.Config) *api.Client {
	apiClient, apiClientAuthLifetimeSeconds = vault.LoginWithAppConfig(appConfig)
	if 0 != apiClientAuthLifetimeSeconds {
		apiClientAuthLifetime = time.Now().Add(time.Second * time.Duration(apiClientAuthLifetimeSeconds))
		glog.V(1).Info("Created api client token expires at ", apiClientAuthLifetime)
	}

	return apiClient
}

func isCurrentApiClientValid() bool {
	return nil != apiClient &&
		(apiClientAuthLifetimeSeconds == 0 || apiClientAuthLifetime.After(time.Now()))
}

func RevokeAuthLease(appConfig config.Config) {
	if nil == apiClient {
		return
	}

	vault.RevokeTokenLease(apiClient)
	apiClient = nil

	data.Clear(appConfig.DataDir)
}
