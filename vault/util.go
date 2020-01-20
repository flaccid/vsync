package vault

import (
	"fmt"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/flaccid/vsync/config"
	"github.com/hashicorp/vault/api"
)

var (
	totalSecrets        int
	totalSecretsFolders int
	childPaths          []string
)

// writeSecret writes a single secret to the provided vault
// a private function that requires providing your vault api client
func writeSecret(v *api.Client, path string, secret *api.Secret) error {
	// TODO: add kv1/2 switching, currently assumes v2
	// only currently matches /secret mount :(
	v2Path := strings.Replace(path, "/secret", "/secret/data", 1)
	data := map[string]interface{}{"data": secret.Data}

	log.Debugf("write the secret to %s with %s", path, secret.Data)

	_, err := v.Logical().Write(v2Path, data)
	if err != nil {
		return err
	}

	return nil
}

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
					log.Fatal("failed to write secret: ", err)
				}
				log.Infof("sync'd %s", newPath)
			}
		}
	}
}

// getMounts returns all mountpoints from the provided vault client
func getMounts(v *api.Client) (mounts map[string]*api.MountOutput, err error) {
	mounts, err = v.Sys().ListMounts()
	return mounts, err
}

// getMount returns the mount provided by an arbitrary secret path
func getMount(v *api.Client, secretPath string) (mount string) {
	mounts, err := getMounts(v)
	if err != nil {
		log.Errorf("error getting mounts: ", err)
	}

	// NOTE: only currently supports one folder depth mountpoints
	mount = strings.Split(secretPath, "/")[1] + "/"
	//log.Debugf("check if %s is in ", secretPath, mounts)
	for k := range mounts {
		if k == mount {
			//log.Debugf("mount %s exists", mount)
			return mount
		}
	}

	return ""
}

// getSecretPaths iterates on a secret path and returns all secret paths found within
func getSecretPaths(v *api.Client, secretPath string) (secretPaths []string, err error) {
	mountPoint := getMount(v, secretPath)
	//log.Debugf("acton on %s in %s", secretPath, mountPoint)
	//log.Debugf("walk %s", path)

	// assumes kv v2 engine
	listPath := normalizeVaultPath(strings.Replace(secretPath, "/"+path.Clean(mountPoint), "/"+path.Clean(mountPoint)+"/metadata", 1))
	//log.Debugf("list path %s", listPath)
	secretsList, err := v.Logical().List(listPath)
	if err != nil {
		log.Fatal(err)
	}
	//log.Debugf("secretsList keys", secretsList.Data["keys"])

	for _, b := range secretsList.Data {
		//log.Debug(a, b)
		for _, p := range b.([]interface{}) {
			if p.(string)[len(p.(string))-1:] == "/" {
				// is a path/folder
				node := normalizeVaultPath(secretPath + "/" + p.(string))
				//log.Debugf("found node %s", node)
				totalSecretsFolders = totalSecretsFolders + 1
				getSecretPaths(v, node)
			} else {
				// is a secret
				p := normalizeVaultPath(secretPath + "/" + p.(string))
				//log.Debugf("found secret %s", p)
				totalSecrets = totalSecrets + 1
				childPaths = append(childPaths, p)
			}
		}
	}

	secretPaths = childPaths
	return secretPaths, err
}
