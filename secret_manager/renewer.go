package secret_manager

import (
	"github.com/golang/glog"
	"github.com/szeber/vault-kubernetes-dotenv-manager/config"
	"github.com/szeber/vault-kubernetes-dotenv-manager/constants"
	"github.com/szeber/vault-kubernetes-dotenv-manager/data"
	"github.com/szeber/vault-kubernetes-dotenv-manager/vault"
	"time"
)

func KeepSecretsAlive(appConfig config.Config, httpPort int) {
	serverHttpPort = httpPort
	startHttpServer()
	savedData := data.Load(appConfig.DataDir)

	renewalCount := 1

	for renewalCount > 0 {
		renewalCount = runRenewal(appConfig, &savedData)
	}
}

func runRenewal(appConfig config.Config, savedData *data.SavedData) int {
	nextProcessingTimestamp := getNextSecretToRenew(savedData)

	if nextProcessingTimestamp == 0 {
		return 0
	}

	nextProcessingTime := time.Unix(nextProcessingTimestamp, 0)

	if nextProcessingTime.After(time.Now()) {
		isAlive = true
		glog.Info("Sleeping until ", nextProcessingTime)
		time.Sleep(nextProcessingTime.Sub(time.Now()))
	}

	renewedSecretCount := renewSecrets(savedData, appConfig)

	data.Save(appConfig.DataDir, *savedData)

	return renewedSecretCount
}

func getNextSecretToRenew(savedData *data.SavedData) int64 {
	var nextProcessingTimestamp int64 = 0

	for _, secretData := range savedData.Secrets {
		if secretData.Renewable && secretData.LeaseDuration > 0 {
			secretNextProcessingTime := savedData.CreationTimestamp + (secretData.LeaseDuration / constants.LifetimeDivisor)
			if nextProcessingTimestamp == 0 || int64(secretNextProcessingTime) < nextProcessingTimestamp {
				nextProcessingTimestamp = int64(savedData.CreationTimestamp + (secretData.LeaseDuration / constants.LifetimeDivisor))
			}
		}
	}

	return nextProcessingTimestamp
}

func renewSecrets(savedData *data.SavedData, appConfig config.Config) int {
	glog.Info("Starting secret lease renewal")

	apiClient := getClient(appConfig)
	renewedSecretCount := 0
	savedData.CreationTimestamp = int(time.Now().UTC().Unix())

	for key, secretData := range savedData.Secrets {
		if secretData.Renewable {
			glog.V(1).Info("Renewing lease " + secretData.LeaseID)
			savedData.Secrets[key] = *vault.RenewSecret(apiClient, secretData)
			glog.V(1).Info("Lease renewed. New validity: ", time.Now().Add(time.Second*time.Duration(int64(savedData.Secrets[key].LeaseDuration))))
			renewedSecretCount = renewedSecretCount + 1
		}
	}

	glog.Infof("Renewed %d leases", renewedSecretCount)

	return renewedSecretCount
}
