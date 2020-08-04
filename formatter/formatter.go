package formatter

import (
	"github.com/golang/glog"
	"github.com/szeber/vault-kubernetes-dotenv-manager/config"
	"github.com/szeber/vault-kubernetes-dotenv-manager/constants"
	"github.com/szeber/vault-kubernetes-dotenv-manager/decoder"
	"github.com/szeber/vault-kubernetes-dotenv-manager/helper"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

func FormatSecret(secretData map[string]string, definition config.SecretDefinition) {
	glog.Info("Writing secret " + definition.Name)
	dec, err := decoder.New(definition)

	if nil != err {
		glog.Exit("Failed to create decoder for secret " + definition.Name)
	}

	switch definition.Format {
	case constants.FormatFile:
		formatFileSecret(secretData, definition, dec)
	case constants.FormatDotenv:
		formatDotenvSecret(secretData, definition, dec)
	default:
		glog.Exit("Invalid format: " + definition.Format)
	}
}

func formatDotenvSecret(secretData map[string]string, definition config.SecretDefinition, dec *decoder.Decoder) {
	headerText := "Secret source: " + definition.Name
	stringToWrite := "\n" + strings.Repeat("#", len(headerText)+4) + "\n# " + headerText + " #\n" + strings.Repeat("#", len(headerText)+4) + "\n"

	if !helper.FileExists(path.Dir(definition.Destination)) {
		err := os.MkdirAll(path.Dir(definition.Destination), 0755)

		if err != nil {
			glog.Exit("Failed to crate desitnation directory for secret "+definition.Name+": ", err)
		}
	}

	for key, value := range mapSecretData(secretData, definition) {
		decodedValue, err := dec.DecodeString(value)

		if nil != err {
			glog.Exitf("Failed to decode value for %s in secret %s: %v", key, definition.Name, err)
		}

		stringToWrite = stringToWrite + key + "=" + strconv.Quote(string(decodedValue)) + "\n"
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

func formatFileSecret(secretData map[string]string, definition config.SecretDefinition, dec *decoder.Decoder) {
	if !helper.FileExists(definition.Destination) {
		err := os.MkdirAll(definition.Destination, 0755)

		if err != nil {
			glog.Exit("Failed to create destination directory for secret " + definition.Name)
		}
	}

	if !helper.IsDir(definition.Destination) {
		glog.Exit("The destination is not a directory for secret " + definition.Name)
	}

	for key, value := range mapSecretData(secretData, definition) {
		decodedValue, err := dec.DecodeString(value)

		if nil != err {
			glog.Exitf("Failed to decode value for %s in secret %s: %v", key, definition.Name, err)
		}

		err = ioutil.WriteFile(definition.Destination+"/"+key, decodedValue, 0644)

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
			glog.Exit("Mapping failed in secret " + definition.Name + ". Key " + value + " doesn't exist in secret data")
		}

		newSecretData[key] = mappedValue
	}

	return newSecretData
}
