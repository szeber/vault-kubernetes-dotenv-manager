package formatter

import (
	"github.com/golang/glog"
	"github.com/szeber/vault-kubernetes-dotenv-manager/config"
	"github.com/szeber/vault-kubernetes-dotenv-manager/constants"
	"github.com/szeber/vault-kubernetes-dotenv-manager/helper"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

func FormatSecret(secretData map[string]string, definition config.SecretDefinition) {
	glog.Info("Writing secret " + definition.Name)
	switch definition.Format {
	case constants.FormatFile:
		formatFileSecret(secretData, definition)
	case constants.FormatDotenv:
		formatDotenvSecret(secretData, definition)
	default:
		glog.Exit("Invalid format: " + definition.Format)
	}
}

func formatDotenvSecret(secretData map[string]string, definition config.SecretDefinition) {
	headerText := "Secret source: " + definition.Name
	stringToWrite := strings.Repeat("#", len(headerText)+4) + "\n# " + headerText + " #\n" + strings.Repeat("#", len(headerText)+4) + "\n"

	if !helper.FileExists(path.Dir(definition.Destination)) {
		err := os.MkdirAll(path.Dir(definition.Destination), 0755)

		if err != nil {
			glog.Exit("Failed to crate desitnation directory for secret "+definition.Name+": ", err)
		}
	}

	for key, value := range mapSecretData(secretData, definition) {
		stringToWrite = stringToWrite + key + "=" + strconv.Quote(value) + "\n"
	}

	f, err := os.OpenFile(definition.Destination, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		glog.Exit("Failed to open destination file for writing for secret "+definition.Name+": ", err)
	}

	defer f.Close()

	_, err = f.WriteString(stringToWrite)

	if err != nil {
		glog.Exit("Failed to write to destination file for secret "+definition.Name+": ", err)
	}
}

func formatFileSecret(secretData map[string]string, definition config.SecretDefinition) {
	if !helper.FileExists(definition.Destination) {
		err := os.MkdirAll(definition.Destination, 0755)

		if err != nil {
			glog.Error("Failed to create destination directory for secret " + definition.Name)
			os.Exit(constants.ExitCodeConfigError)
		}
	}

	if !helper.IsDir(definition.Destination) {
		glog.Error("The destination is not a directory for secret " + definition.Name)
		os.Exit(constants.ExitCodeConfigError)
	}

	for key, value := range mapSecretData(secretData, definition) {
		err := ioutil.WriteFile(definition.Destination+"/"+key, []byte(value), 0644)

		if err != nil {
			glog.Exit("Failed to write file "+key+" for secret "+definition.Name+": ", err)
		}
	}
}

func mapSecretData(secretData map[string]string, definition config.SecretDefinition) map[string]string {
	if len(definition.Mapping) == 0 {
		return secretData
	}

	newSecretData := map[string]string{}

	for key, value := range definition.Mapping {
		mappedValue, ok := secretData[value]

		if !ok {
			glog.Error("Mapping failed in secret " + definition.Name + ". Key " + value + " doesn't exist in secret data")
			os.Exit(constants.ExitCodeConfigError)
		}

		newSecretData[key] = mappedValue
	}

	return newSecretData
}
