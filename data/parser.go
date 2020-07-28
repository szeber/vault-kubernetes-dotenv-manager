package data

import (
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	"github.com/szeber/vault-kubernetes-dotenv-manager/helper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type SavedData struct {
	CreationTimestamp int          `yaml:"creationTimestamp"`
	LoginToken        string       `yaml:"LoginToken"`
	AuthLeaseDuration int          `yaml:"AuthLeaseDuration"`
	Secrets           []api.Secret `yaml:"secrets"`
}

func Load(basePath string) SavedData {
	filePath := getFilePath(basePath)
	savedData := SavedData{}

	if !helper.FileExists(filePath) {
		glog.Exit("Data file does not exist")
	}

	yamlContents, err := ioutil.ReadFile(filePath)

	if err != nil {
		glog.Exit("Failed to load the data file: ", err)
	}

	err = yaml.Unmarshal(yamlContents, &savedData)

	if err != nil {
		glog.Exit("Failed to parse the data file as YAML. ", err)
	}

	return savedData
}

func Save(basePath string, data SavedData) {
	filePath := getFilePath(basePath)
	yamlContents, err := yaml.Marshal(data)

	if err != nil {
		glog.Exit("Failed to create yaml data: ", err)
	}

	err = ioutil.WriteFile(filePath, yamlContents, 0644)

	if err != nil {
		glog.Exit("Failed to write data file: ", err)
	}
}

func Clear(basePath string) {
	filePath := getFilePath(basePath)

	if helper.FileExists(filePath) {
		err := os.Remove(filePath)

		if err != nil {
			glog.Exit("Failed to delete data file: ", err)
		}
	}
}

func getFilePath(basePath string) string {
	if "" == basePath {
		glog.Exit("Empty basedir")
	}
	return basePath + "/data.yaml"
}
