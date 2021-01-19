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

	if err := renewSecrets(savedData, appConfig); err != nil {
		if time.Now().After(savedData.GetTimeOfShortestExpiration().Add(-5 * time.Second)) {
			glog.Warning("Error while renewing secrets. Sleeping for 5 seconds before trying again")
			time.Sleep(5 * time.Second)
		} else {
			glog.Exit("Failed to renew the secrets, giving up: ", err)
		}
	}

	data.Save(appConfig.DataDir, *savedData)
}

func getNextSecretToRenew(savedData *data.SavedData) int64 {
	return int64(savedData.CreationTimestamp + (savedData.GetShortestExpirationSeconds() / constants.LifetimeDivisor))
}

func renewSecrets(savedData *data.SavedData, appConfig config.Config) error {
	glog.Info("Starting lease renewals")

	apiClient := getClient(appConfig, savedData.LoginToken)
	newCreationTimestamp := int(time.Now().UTC().Unix())
	apiClientAuthLifetimeSeconds, err := vault.RenewTokenLease(apiClient)

	if nil != err {
		return err
	}

	apiClientAuthLifetime = time.Now().Add(time.Second * time.Duration(apiClientAuthLifetimeSeconds))
	renewedSecretCount := 0

	glog.Info("Renewed auth token lease. New expiration: ", apiClientAuthLifetime)

	for key, secretData := range savedData.Secrets {
		if secretData.Renewable {
			glog.V(1).Info("Renewing lease " + secretData.LeaseID)
			newSecret, err := vault.RenewLease(apiClient, secretData)

			if nil != err {
				return err
			}

			savedData.Secrets[key] = *newSecret

			glog.V(1).Info(
				"Lease renewed. New validity: ",
				time.Now().Add(time.Second*time.Duration(int64(savedData.Secrets[key].LeaseDuration))),
			)
			renewedSecretCount = renewedSecretCount + 1
		}
	}

	savedData.CreationTimestamp = newCreationTimestamp
	glog.Infof("Renewed %d secret leases", renewedSecretCount)

	return nil
}
