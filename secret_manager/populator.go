package secret_manager

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	"github.com/szeber/vault-kubernetes-dotenv-manager/config"
	"github.com/szeber/vault-kubernetes-dotenv-manager/constants"
	"github.com/szeber/vault-kubernetes-dotenv-manager/data"
	"github.com/szeber/vault-kubernetes-dotenv-manager/formatter"
	"github.com/szeber/vault-kubernetes-dotenv-manager/helper"
	"io/ioutil"
	"path"
	"time"
)

func PopulateSecrets(appConfig config.Config) {
	apiClient := getClient(appConfig, "")

	glog.Info("Starting secret population")
	data.Clear(appConfig.DataDir)

	dataToSave := data.SavedData{
		CreationTimestamp: int(time.Now().UTC().Unix()),
		LoginToken:        apiClient.Token(),
		AuthLeaseDuration: apiClientAuthLifetimeSeconds,
	}

	for _, definition := range appConfig.Secrets {
		glog.Info("Populating secret " + definition.Name)
		var secretData map[string]string

		switch definition.Origin {
		case constants.OriginVault:
			secretData = getSecretFromVault(apiClient, definition, &dataToSave)
		case constants.OriginFile:
			secretData = getSecretFromFile(definition)
		default:
			panic("Invalid origin: " + definition.Origin)
		}

		formatter.FormatSecret(secretData, definition)
	}

	data.Save(appConfig.DataDir, dataToSave)
	glog.Info("Finished secret population")
}

func getSecretFromVault(apiClient *api.Client, defintion config.SecretDefinition, savedData *data.SavedData) map[string]string {
	response, err := apiClient.Logical().Read(defintion.Source)

	if nil != err {
		glog.Error("Failed to load secret "+defintion.Source+" from Vault: ", err)
	}

	secretData := map[string]string{}

	for key, value := range getDataForSubKey(response.Data, defintion.SecretBaseKey) {
		secretData[key] = fmt.Sprintf("%v", value)
	}

	response.Data = nil

	savedData.Secrets = append(savedData.Secrets, *response)

	return secretData
}

func getDataForSubKey(sourceData map[string]interface{}, key string) map[string]interface{} {
	if key == "" {
		return sourceData
	}

	subKeyData, ok := sourceData[key]

	if !ok {
		glog.Exit("Failed to get data from secret under base key " + key)
	}

	switch v := subKeyData.(type) {
	case map[string]interface{}:
		return v
	default:
		glog.Exit("Invalid type for secret data")
	}

	return nil
}

func getSecretFromFile(definition config.SecretDefinition) map[string]string {
	if !helper.FileExists(definition.Source) {
		glog.Exit("Source doesn't exist for secret: " + definition.Source)
	}

	fileData, err := ioutil.ReadFile(definition.Source)

	if err != nil {
		glog.Exit("Failed to read secret from source file: " + definition.Source)
	}

	return map[string]string{
		path.Base(definition.Source): string(fileData),
	}
}
