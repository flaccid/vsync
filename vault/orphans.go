package vault

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/flaccid/vsync/config"
)

// RemoveOprhans removes secret paths in the destination vault that no longer exist in the source vault
func (v *Client) RemoveOrphans(appConfig *config.AppConfig, path string) (secretPaths []string, err error) {
	log.Debugf("remove orphans from %s", path)

	var orphans []string
	secretPaths, err = getSecretPaths(appConfig.Destination.Client, path)

	// for each secret found in the destination,
	// see if it exists in the source and remove if not found
	for _, secretPath := range secretPaths {
		secret, err := v.ReadSecret(appConfig, secretPath, false)
		if err != nil {
			log.Error(err)
			orphans = append(orphans, secretPath)
		} else {
			if secret.Data != nil {
				log.Debugf("%s: ", secretPath, secret)
			}
		}
	}

	secretPaths = orphans
	log.Debugf("secrets to remove: ", secretPaths)

	// remove the orphans
	for _, orphan := range orphans {
		if appConfig.DryRun != true {
			log.Info("remove " + orphan)
			// assume kv2
			err := v.DeleteSecret(appConfig, strings.Replace(orphan, "/secret", "/secret/data", 1), true)
			if err != nil {
				log.Errorf("failed to delete secret: %s", err)
			}
		} else {
			log.Infof("dry run, skipping actual removal of %s", orphan)
		}
	}

	return secretPaths, err
}
