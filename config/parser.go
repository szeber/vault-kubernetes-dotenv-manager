package config

import (
	"github.com/golang/glog"
	"github.com/szeber/vault-kubernetes-dotenv-manager/constants"
	"github.com/szeber/vault-kubernetes-dotenv-manager/helper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Config struct {
	DataDir            string             `yaml:"dataDir"`
	VaultUrl           string             `yaml:"vaultUrl"`
	TokenPath          string             `yaml:"tokenPath"`
	Namespace          string             `yaml:"namespace"`
	Role               string             `yaml:"role"`
	VaulAuthMethodPath string             `yaml:"vaultAuthMethodPath"`
	Secrets            []SecretDefinition `yaml:"secrets"`
}

type SecretDefinition struct {
	Name          string            `yaml:"name"`
	Origin        string            `yaml:"origin"`
	Source        string            `yaml:"source"`
	Destination   string            `yaml:"destination"`
	Format        string            `yaml:"format"`
	SecretBaseKey string            `yaml:"secretBaseKey"`
	Mapping       map[string]string `yaml:"mapping"`
}

func LoadConfig(configPath string) Config {
	glog.V(1).Info("Loading config from " + configPath)

	config := Config{}

	if !helper.FileExists(configPath) {
		glog.Error("Config file does not exist")
		os.Exit(constants.ExitCodeConfigParseError)
	}

	yamlContents, err := ioutil.ReadFile(configPath)

	if err != nil {
		glog.Error("Failed to load the config file: ", err)
		os.Exit(constants.ExitCodeConfigParseError)
	}

	err = yaml.Unmarshal(yamlContents, &config)

	if err != nil {
		glog.Error("Failed to parse the config file as YAML. ", err)
		os.Exit(constants.ExitCodeConfigParseError)
	}

	populateDefaults(&config)

	validateConfig(config)

	glog.V(1).Info("Finished loading the config file")

	return config
}

func validateConfig(config Config) {
	errors := []string{}

	prepareAndValidateDataDir(config.DataDir, &errors)

	if "" == config.Namespace {
		config.Namespace = "default"
	}

	if "" == config.VaultUrl {
		errors = append(errors, "No Vault URL set")
	}

	if "" == config.Role {
		errors = append(errors, "No role set")
	}

	if "" == config.VaulAuthMethodPath {
		errors = append(errors, "No Vault auth method path set")
	}

	for i := range config.Secrets {
		validateSecret(&config.Secrets[i], i, &errors)
	}

	if !helper.FileExists(config.TokenPath) {
		errors = append(errors, "Token file does not exist at "+config.TokenPath)
	}

	if len(errors) > 0 {
		glog.Error("Validation failed for the config file")
		for i := range errors {
			glog.Error(errors[i])
		}
		os.Exit(constants.ExitCodeConfigParseError)
	}
}

func prepareAndValidateDataDir(dataDir string, errors *[]string) {
	if "" == dataDir {
		*errors = append(*errors, "No data directory defined")
		return
	}

	if !helper.FileExists(dataDir) {
		err := os.MkdirAll(dataDir, 0755)

		if err != nil {
			*errors = append(*errors, "Failed to create data directory. Error: "+err.Error())
		}
	} else if !helper.IsDir(dataDir) {
		*errors = append(*errors, "The data directory is not a directory")
	}
}

func validateSecret(secret *SecretDefinition, i int, errors *[]string) {
	if "" == secret.Name {
		*errors = append(*errors, "No name for secret #"+string(i))
	}

	if "" == secret.Source {
		*errors = append(*errors, "No source for secret #"+string(i))
	}

	if "" == secret.Destination {
		*errors = append(*errors, "No destination for secret #"+string(i))
	}

	if "" == secret.Origin {
		secret.Origin = constants.OriginVault
	} else if !helper.StringInSlice(constants.ValidOrigins[:], secret.Origin) {
		*errors = append(*errors, "Invalid origin for secret #"+string(i)+": "+secret.Origin)
	}

	if !helper.StringInSlice(constants.ValidFormats[:], secret.Format) {
		*errors = append(*errors, "Invalid format #"+string(i)+": "+secret.Format)
	}
}

func populateDefaults(config *Config) {
	if "" == config.TokenPath {
		config.TokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	}
}
