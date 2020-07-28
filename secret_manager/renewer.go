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

	for {
		runRenewal(appConfig, &savedData)
	}
}

func runRenewal(appConfig config.Config, savedData *data.SavedData) {
	getClient(appConfig, savedData.LoginToken)
	nextProcessingTimestamp := getNextSecretToRenew(savedData)

	nextProcessingTime := time.Unix(nextProcessingTimestamp, 0)

	if nextProcessingTime.After(time.Now()) {
		isAlive = true
		glog.Info("Sleeping until ", nextProcessingTime)
		time.Sleep(nextProcessingTime.Sub(time.Now()))
	}

	renewSecrets(savedData, appConfig)

	data.Save(appConfig.DataDir, *savedData)
}

func getNextSecretToRenew(savedData *data.SavedData) int64 {
	nextProcessingTimestamp := int64(savedData.CreationTimestamp + (savedData.AuthLeaseDuration / constants.LifetimeDivisor))

	for _, secretData := range savedData.Secrets {
		if secretData.Renewable && secretData.LeaseDuration > 0 {
			secretNextProcessingTime := savedData.CreationTimestamp + (secretData.LeaseDuration / constants.LifetimeDivisor)
			if int64(secretNextProcessingTime) < nextProcessingTimestamp {
				nextProcessingTimestamp = int64(savedData.CreationTimestamp + (secretData.LeaseDuration / constants.LifetimeDivisor))
			}
		}
	}

	return nextProcessingTimestamp
}

func renewSecrets(savedData *data.SavedData, appConfig config.Config) {
	glog.Info("Starting lease renewals")

	apiClient := getClient(appConfig, savedData.LoginToken)
	savedData.CreationTimestamp = int(time.Now().UTC().Unix())
	apiClientAuthLifetimeSeconds = vault.RenewTokenLease(apiClient)
	apiClientAuthLifetime = time.Now().Add(time.Second * time.Duration(apiClientAuthLifetimeSeconds))
	renewedSecretCount := 0

	glog.Info("Renewed auth token lease. New expiration: ", apiClientAuthLifetime)

	for key, secretData := range savedData.Secrets {
		if secretData.Renewable {
			glog.V(1).Info("Renewing lease " + secretData.LeaseID)
			savedData.Secrets[key] = *vault.RenewLease(apiClient, secretData)
			glog.V(1).Info(
				"Lease renewed. New validity: ",
				time.Now().Add(time.Second*time.Duration(int64(savedData.Secrets[key].LeaseDuration))),
			)
			renewedSecretCount = renewedSecretCount + 1
		}
	}

	glog.Infof("Renewed %d secret leases", renewedSecretCount)
}
