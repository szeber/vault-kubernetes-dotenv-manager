package vault

import (
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
)

func RenewLease(apiClient *api.Client, secret api.Secret) *api.Secret {
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

func RevokeTokenLease(apiClient *api.Client) {
	glog.V(1).Info("Revoking token lease")
	request := apiClient.NewRequest("POST", "/v1/auth/token/revoke-self")
	_, err := apiClient.RawRequest(request)

	if err != nil {
		glog.Error("Failed to revoke lease: ", err)
	}

	glog.Info("Token lease revocation complete")
}

func RenewTokenLease(apiClient *api.Client) int {
	glog.V(1).Info("Renewing token lease")
	request := apiClient.NewRequest("POST", "/v1/auth/token/renew-self")
	result, err := apiClient.RawRequest(request)

	if err != nil {
		glog.Exit("Failed to renew lease: ", err)
	}

	secret, err := api.ParseSecret(result.Body)

	if err != nil {
		glog.Exit("Failed to create secret while renewing lease: ", err)
	}

	return secret.Auth.LeaseDuration
}
