package secret_manager

import (
	"github.com/hashicorp/vault/api"
	"github.com/szeber/vault-kubernetes-dotenv-manager/config"
	"github.com/szeber/vault-kubernetes-dotenv-manager/vault"
	"time"
)

var apiClientAuthLifetimeSeconds int
var apiClientAuthLifetime time.Time
var apiClient *api.Client

func getClient(appConfig config.Config) *api.Client {
	if nil == apiClient {
		return makeClient(appConfig)
	}

	if apiClientAuthLifetimeSeconds == 0 || apiClientAuthLifetime.After(time.Now().Add(5*time.Second)) {
		return apiClient
	}

	return makeClient(appConfig)
}

func makeClient(appConfig config.Config) *api.Client {
	apiClient, apiClientAuthLifetimeSeconds = vault.LoginWithAppConfig(appConfig)
	if 0 != apiClientAuthLifetimeSeconds {
		apiClientAuthLifetime = time.Now().Add(time.Second * time.Duration(apiClientAuthLifetimeSeconds))
	}

	return apiClient
}
