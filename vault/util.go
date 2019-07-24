package vault

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/flaccid/vsync/config"
)

// normalizeVaultPath takes out possible double slashes
func normalizeVaultPath(path string) (newPath string) {
	return strings.Replace(path, "//", "/", -1)
}

// walkNode iterates on a secret path that appears to be a folder
func walkNode(v *Client, path string) {
	path = normalizeVaultPath(path)
	log.Debug("walk ", path)

	secretsList, err := v.client.Logical().List(path)
	if err != nil {
		log.Fatal(err)
	}

	for a, b := range secretsList.Data {
		log.Debug(a, b)
		for _, p := range b.([]interface{}) {
			if p.(string)[len(p.(string))-1:] == "/" {
				node := p.(string)
				walkNode(v, path+"/"+node)
			} else {
				secret, err := v.ReadSecret(path + "/" + p.(string))
				if err != nil {
					log.Panic(err)
				}
				log.Info(secret)
			}
		}
	}
}

// syncNode iterates on a secret path on source to sync to destination
func syncNode(v *Client, appConfig *config.AppConfig, path string) {
	// get the secrets list at the entrypoint path
	secretsList, err := v.client.Logical().List(path)
	if err != nil {
		log.Fatal(err)
	}

	for a, b := range secretsList.Data {
		log.Debug(a, b)
		for _, p := range b.([]interface{}) {
			if p.(string)[len(p.(string))-1:] == "/" {
				node := p.(string)
				syncNode(v, appConfig, path+"/"+node)
			} else {
				newPath := normalizeVaultPath(path + "/" + p.(string))

				// get the secret from the source
				secret, err := v.ReadSecret(newPath)
				if err != nil {
					log.Panic(err)
				}
				// WARNING: insecure
				log.Debug(secret, err)

				// sync the secret to the destination vault
				err = writeSecret(appConfig.Destination.Client, newPath, secret)
				if err != nil {
					log.Fatal("failed to write secret", err)
				}
				log.Infof("sync'd %s", newPath)
			}
		}
	}
}
