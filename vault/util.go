package vault

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/flaccid/vsync/config"
	"github.com/hashicorp/vault/api"
)

// returns client based on appconfig provided
func getClient(appConfig *config.AppConfig, destinationVault bool) (client *api.Client) {
	if destinationVault {
		return appConfig.Destination.Client
	} else {
		return appConfig.Source.Client
	}
}

// normalizeVaultPath takes out possible double slashes
func normalizeVaultPath(path string) (newPath string) {
	return strings.Replace(path, "//", "/", -1)
}

// dumpNode iterates on a secret path and dumps all recursiviely
func dumpNode(v *api.Client, path string) {
	log.Debug("api client ", v)
	path = normalizeVaultPath(path)
	log.Debugf("walk %s", path)

	secretsList, err := v.Logical().List(path)
	if err != nil {
		log.Fatal(err)
	}

	for a, b := range secretsList.Data {
		log.Debug(a, b)
		for _, p := range b.([]interface{}) {
			if p.(string)[len(p.(string))-1:] == "/" {
				node := p.(string)
				dumpNode(v, path+"/"+node)
			} else {
				p := normalizeVaultPath(path + "/" + p.(string))
				secret, err := v.Logical().Read(p)
				if err != nil {
					log.Panic(err)
				}
				fmt.Printf("    %s:\n", p)
				fmt.Println(secret.Data)
			}
		}
	}
}

// syncNode iterates on a secret path on source to sync to destination
func syncNode(v *Client, appConfig *config.AppConfig, path string) {
	// get the secrets list at the entrypoint path
	secretsList, err := appConfig.Source.Client.Logical().List(path)
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
				secret, err := v.ReadSecret(appConfig, newPath, false)
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
