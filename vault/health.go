package vault

import (
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/flaccid/vsync/config"
)

// HealthCheck performs a health check on the vault server
func (v *Client) HealthCheck(appConfig *config.AppConfig, destinationVault bool) (status string, err error) {
	log.Info("checking health")

	result, err := v.Request(appConfig, destinationVault, "GET", "/sys/health", "")
	if err != nil {
		log.Error(err)
		log.Fatalf("error getting health status: %s", err)
	}

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		log.Fatalf("error reading http body: %s", err)
	}

	return string(body), err
}
