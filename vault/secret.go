package vault

import (
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
)

func RenewSecret(apiClient *api.Client, secret api.Secret) *api.Secret {
	request := apiClient.NewRequest("PUT", "/v1/sys/leases/renew")
	body := map[string]interface{}{
		"lease_id":  secret.LeaseID,
		"increment": secret.LeaseDuration,
	}
	err := request.SetJSONBody(body)

	if err != nil {
		glog.Exit("Failed to set body for lease renewal request: ", err)
	}

	response, err := apiClient.RawRequest(request)

	if err != nil {
		glog.Exit("Failed to renew lease for secret: ", err)
	}

	createdSecret, err := api.ParseSecret(response.Body)

	if err != nil {
		glog.Exit("Failed to create secret while renewing lease: ", err)
	}

	return createdSecret
}
