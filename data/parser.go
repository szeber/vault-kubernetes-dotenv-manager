package data

import (
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	"github.com/szeber/vault-kubernetes-dotenv-manager/constants"
	"github.com/szeber/vault-kubernetes-dotenv-manager/helper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type SavedData struct {
	CreationTimestamp int          `yaml:"creationTimestamp"`
	Secrets           []api.Secret `yaml:"secrets"`
}

func Load(basePath string) SavedData {
	filePath := getFilePath(basePath)
	savedData := SavedData{}

	if !helper.FileExists(filePath) {
		glog.Error("Data file does not exist")
		os.Exit(constants.ExitCodeConfigParseError)
	}

	yamlContents, err := ioutil.ReadFile(filePath)

	if err != nil {
		glog.Error("Failed to load the data file: ", err)
		os.Exit(constants.ExitCodeConfigParseError)
	}

	err = yaml.Unmarshal(yamlContents, &savedData)

	if err != nil {
		glog.Error("Failed to parse the data file as YAML. ", err)
		os.Exit(constants.ExitCodeConfigParseError)
	}

	return savedData
}

func Save(basePath string, data SavedData) {
	filePath := getFilePath(basePath)
	yamlContents, err := yaml.Marshal(data)

	if err != nil {
		glog.Error("Failed to create yaml data: ", err)
		os.Exit(constants.ExitCodeConfigParseError)
	}

	err = ioutil.WriteFile(filePath, yamlContents, 0644)

	if err != nil {
		glog.Error("Failed to write data file: ", err)
		os.Exit(constants.ExitCodeConfigParseError)
	}
}

func Clear(basePath string) {
	filePath := getFilePath(basePath)

	if helper.FileExists(filePath) {
		err := os.Remove(filePath)

		if err != nil {
			glog.Error("Failed to delete data file: ", err)
			os.Exit(constants.ExitCodeConfigParseError)
		}
	}
}

func getFilePath(basePath string) string {
	if "" == basePath {
		glog.Error("Empty basedir")
		os.Exit(constants.ExitCodeConfigParseError)
	}
	return basePath + "/data.yaml"
}
