package vault

import (
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	"github.com/szeber/vault-kubernetes-dotenv-manager/config"
	"io/ioutil"
	"net/http"
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
	apiClient, leaseDuration, err := Login(getAuthConfigFromAppConfig(appConfig))

	if nil != err {
		glog.Exit("Failed to log in to Vault")
	}

	return apiClient, leaseDuration
}

func Login(authConfig AuthConfig) (*api.Client, int, error) {
	apiClient, err := getApiClient(authConfig)

	if err != nil {
		return nil, 0, err
	}

	result, err := sendLoginRequest(apiClient, authConfig)

	if nil != err {
		return nil, 0, err
	}

	glog.V(1).Infof("Login token duration: %d", result.Auth.LeaseDuration)
	glog.Info("Login successful")

	apiClient.SetToken(result.Auth.ClientToken)

	return apiClient, result.Auth.LeaseDuration, nil
}

func sendLoginRequest(apiClient *api.Client, authConfig AuthConfig) (*api.Secret, error) {
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

	req := apiClient.NewRequest("POST", loginPath)
	err := req.SetJSONBody(body)

	if err != nil {
		glog.Error("ERROR: Failed to set json body: ", err)
	}

	resp, err := apiClient.RawRequest(req)
	if err != nil {
		glog.Errorf("ERROR: failed to login with Vault %s", req.URL.String())
		glog.Error(err)
		return nil, err
	}

	if respErr := resp.Error(); respErr != nil {
		glog.Errorf("ERROR: api error: %v", respErr)
		return nil, respErr
	}

	var result api.Secret
	if err := resp.DecodeJSON(&result); err != nil {
		glog.Errorf("ERROR: failed to decode JSON response: %v", err)
		return nil, err
	}

	return &result, nil
}

func getApiClient(authConfig AuthConfig) (*api.Client, error) {
	glog.Infof("Connecting to Vault at %s", authConfig.Url)

	httpClient := buildHTTPClient(authConfig.Url)

	apiConfig := &api.Config{
		Address:    authConfig.Url,
		HttpClient: httpClient,
	}

	apiClient, err := api.NewClient(apiConfig)

	if err != nil {
		glog.Errorf("ERROR: failed to connect to Vault at %s: %v", authConfig.Url, err)
		return nil, err
	}

	if authConfig.Namespace != "" {
		apiClient.SetNamespace(authConfig.Namespace)
	}

	return apiClient, nil
}

func buildHTTPClient(url string) *http.Client {

	if strings.HasPrefix(url, "http://") {
		return http.DefaultClient
	}
	httpClient := &http.Client{}
	return httpClient
}

func getAuthConfigFromAppConfig(appConfig config.Config) AuthConfig {
	authToken, err := ioutil.ReadFile(appConfig.TokenPath)

	if err != nil {
		glog.Exit("Failed to load the service account token: ", err)
	}

	return AuthConfig{
		Url:                 appConfig.VaultUrl,
		Namespace:           appConfig.Namespace,
		KubeAuthRole:        appConfig.Role,
		KubeAuthPath:        appConfig.VaulAuthMethodPath,
		ServiceAccountToken: string(authToken),
	}
}

func GetClientWithToken(appConfig config.Config, token string) (*api.Client, error) {
	authConfig := getAuthConfigFromAppConfig(appConfig)

	apiClient, err := getApiClient(authConfig)

	if err != nil {
		return nil, err
	}

	apiClient.SetToken(token)

	return apiClient, nil
}
