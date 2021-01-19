package vault

import (
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
)

func RenewLease(apiClient *api.Client, secret api.Secret) (*api.Secret, error) {
	request := apiClient.NewRequest("PUT", "/v1/sys/leases/renew")
	body := map[string]interface{}{
		"lease_id":  secret.LeaseID,
		"increment": secret.LeaseDuration,
	}
	err := request.SetJSONBody(body)

	if err != nil {
		glog.Error("Failed to set body for lease renewal request: ", err)
		return nil, err
	}

	response, err := apiClient.RawRequest(request)

	if err != nil {
		glog.Error("Failed to renew lease for secret: ", err)
		return nil, err
	}

	createdSecret, err := api.ParseSecret(response.Body)

	if err != nil {
		glog.Error("Failed to create secret while renewing lease: ", err)
		return nil, err
	}

	return createdSecret, nil
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

func RenewTokenLease(apiClient *api.Client) (int, error) {
	glog.V(1).Info("Renewing token lease")
	request := apiClient.NewRequest("POST", "/v1/auth/token/renew-self")
	result, err := apiClient.RawRequest(request)

	if err != nil {
		glog.Error("Failed to renew lease: ", err)
		return 0, err
	}

	secret, err := api.ParseSecret(result.Body)

	if err != nil {
		glog.Error("Failed to create secret while renewing lease: ", err)
		return 0, err
	}

	return secret.Auth.LeaseDuration, nil
}
