package vault

import (
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	"github.com/szeber/vault-kubernetes-dotenv-manager/config"
	"github.com/szeber/vault-kubernetes-dotenv-manager/constants"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

type AuthConfig struct {
	Url                 string
	Namespace           string
	KubeAuthRole        string
	KubeAuthPath        string
	ServiceAccountToken string
}

func LoginWithAppConfig(appConfig config.Config) (*api.Client, int) {
	authToken, err := ioutil.ReadFile(appConfig.TokenPath)

	if err != nil {
		glog.Error("Failed to load the service account token: ", err)
		os.Exit(constants.ExitCodeConfigError)
	}

	apiClient, leaseDuration, err := Login(AuthConfig{
		Url:                 appConfig.VaultUrl,
		Namespace:           appConfig.Namespace,
		KubeAuthRole:        appConfig.Role,
		KubeAuthPath:        appConfig.VaulAuthMethodPath,
		ServiceAccountToken: string(authToken),
	})

	if nil != err {
		glog.Error("Failed to log in to Vault")
		os.Exit(constants.ExitCodeConfigError)
	}

	return apiClient, leaseDuration
}

func Login(authConfig AuthConfig) (*api.Client, int, error) {
	glog.Infof("Connecting to Vault at %s", authConfig.Url)

	httpClient := buildHTTPClient(authConfig.Url)

	apiConfig := &api.Config{
		Address:    authConfig.Url,
		HttpClient: httpClient,
	}

	client, err := api.NewClient(apiConfig)

	if err != nil {
		glog.Errorf("ERROR: failed to connect to Vault at %s: %v", authConfig.Url, err)
		return nil, 0, err
	}

	if authConfig.Namespace != "" {
		client.SetNamespace(authConfig.Namespace)
	}
	body := map[string]interface{}{
		"role": authConfig.KubeAuthRole,
		"jwt":  authConfig.ServiceAccountToken,
	}

	loginPath := "/v1/auth/" + authConfig.KubeAuthPath + "/login"
	loginPath = path.Clean(loginPath)
	glog.V(1).Infof(
		"Vault login using path %s role %s jwt [%d bytes]",
		loginPath,
		authConfig.KubeAuthRole,
		len(authConfig.ServiceAccountToken),
	)

	req := client.NewRequest("POST", loginPath)
	err = req.SetJSONBody(body)

	if err != nil {
		glog.Error("ERROR: Failed to set json body: ", err)
	}

	resp, err := client.RawRequest(req)
	if err != nil {
		glog.Errorf("ERROR: failed to login with Vault %s", req.URL.String())
		glog.Error(err)
		return nil, 0, err
	}

	if respErr := resp.Error(); respErr != nil {
		glog.Errorf("ERROR: api error: %v", respErr)
		return nil, 0, respErr
	}

	var result api.Secret
	if err := resp.DecodeJSON(&result); err != nil {
		glog.Errorf("ERROR: failed to decode JSON response: %v", err)
		return nil, 0, err
	}

	glog.V(1).Infof("Login token duration; %d", result.LeaseDuration)
	glog.Info("Login successful")

	client.SetToken(result.Auth.ClientToken)

	leaseDuration := result.LeaseDuration
	return client, leaseDuration, nil
}

func buildHTTPClient(url string) *http.Client {

	if strings.HasPrefix(url, "http://") {
		return http.DefaultClient
	}
	httpClient := &http.Client{}
	return httpClient
}
