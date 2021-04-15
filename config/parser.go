package config

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/szeber/vault-kubernetes-dotenv-manager/constants"
	"github.com/szeber/vault-kubernetes-dotenv-manager/helper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Config struct {
	DataDir               string             `yaml:"dataDir"`
	VaultUrl              string             `yaml:"vaultUrl"`
	TokenPath             string             `yaml:"tokenPath"`
	Namespace             string             `yaml:"namespace"`
	Role                  string             `yaml:"role"`
	VaulAuthMethodPath    string             `yaml:"vaultAuthMethodPath"`
	RevokeAuthLeaseOnQuit bool               `yaml:"revokeAuthLeaseOnQuit"`
	Secrets               []SecretDefinition `yaml:"secrets"`
}

type SecretDefinition struct {
	Name          string            `yaml:"name"`
	Origin        string            `yaml:"origin"`
	Source        string            `yaml:"source"`
	Destination   string            `yaml:"destination"`
	Format        string            `yaml:"format"`
	SecretBaseKey string            `yaml:"secretBaseKey"`
	Mapping       map[string]string `yaml:"mapping"`
	Decoders      []string          `yaml:"decoders"`
}

func LoadConfig(configPath string) Config {
	glog.V(1).Info("Loading config from " + configPath)

	config := Config{}

	if !helper.FileExists(configPath) {
		glog.Exit("Config file does not exist")
	}

	yamlContents, err := ioutil.ReadFile(configPath)

	if err != nil {
		glog.Exit("Failed to load the config file: ", err)
	}

	err = yaml.Unmarshal(yamlContents, &config)

	if err != nil {
		glog.Exit("Failed to parse the config file as YAML. ", err)
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
		os.Exit(1)
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
	if "" == secret.Origin {
		secret.Origin = constants.OriginVault
	} else if !helper.StringInSlice(constants.ValidOrigins[:], secret.Origin) {
		*errors = append(*errors, fmt.Sprintf("Invalid origin for secret #%d: %s", i, secret.Origin))
	}

	if "" == secret.Name {
		*errors = append(*errors, fmt.Sprintf("No name for secret #%d", i))
	}

	if "" == secret.Source && secret.Origin != constants.OriginToken {
		*errors = append(*errors, fmt.Sprintf("No source for secret #%d", i))
	}

	if "" == secret.Destination {
		*errors = append(*errors, fmt.Sprintf("No destination for secret #%d", i))
	}

	if !helper.StringInSlice(constants.ValidFormats[:], secret.Format) {
		*errors = append(*errors, fmt.Sprintf("Invalid format #%d: %s", i, secret.Format))
	}

	for _, decoder := range secret.Decoders {
		if !helper.StringInSlice(constants.ValidDecoders[:], decoder) {
			*errors = append(*errors, fmt.Sprintf("Invalid decoder #%d: %s", i, decoder))
		}
	}
}

func populateDefaults(config *Config) {
	if "" == config.TokenPath {
		config.TokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	}
}
