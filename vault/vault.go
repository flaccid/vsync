package vault

import (
	"errors"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/flaccid/vsync/config"
	"github.com/hashicorp/vault/api"
)

const (
	apiVersion = "v1"
)

var (
	client     *api.Client
	entryPoint string
	secretPath string
)

// GetClient returns the underlining client
func (v *Client) GetClient() *api.Client {
	return v.Client
}

// ReadSecret reads a single secret from the vault
func (v *Client) ReadSecret(appConfig *config.AppConfig, path string, destinationVault bool) (*api.Secret, error) {
	client = getClient(appConfig, destinationVault)

	// read secret depending on secret engine version
	if engineType(client, path) == "kv" {
		path = strings.Replace(path, "/secret", "/secret/data", 1)
	}

	secret, err := client.Logical().Read(path)
	if secret == nil {
		return nil, errors.New("no secret found in " + path)
	}

	return secret, err
}

// WriteSecret writes a single secret to the vault
func (v *Client) WriteSecret(appConfig *config.AppConfig, secret *Secret, destinationVault bool) error {
	log.Debugf("write the secret to %s with %s", secret.Path, secret.Values)
	client = getClient(appConfig, destinationVault)

	// check if the secret data already exists and is the same
	existingSecret, err := v.ReadSecret(appConfig, secret.Path, destinationVault)
	if err != nil {
		log.Debug("secret may not exist: %s", err)
	}
	log.Debugf("existing secret: %s", existingSecret)

	// WARNING: insecure!
	log.Debugf("comparing secret data: cmd[%s] and existing[%s]", secret.Values, existingSecret.Data)

	// when the secret doesn't exist or the values are not the same
	// TODO: below assumes the existing secret is of kv2!
	if existingSecret == nil || ! reflect.DeepEqual(secret.Values, existingSecret.Data["data"]) {
		log.Debug("secret appears to need sync")
		err := writeSecret(client, secret.Path, secret.Values)
		if err != nil {
			return err
		}
		log.Infof("secret written to %s", secret.Path)
	} else {
		log.Info("secret appears to be up to date, not writing")
	}

	return nil
}

// DeleteSecret delets a single secret from the vault
func (v *Client) DeleteSecret(appConfig *config.AppConfig, secretPath string, destinationVault bool) error {
	log.Debugf("delete the secret %s", secretPath)

	client = getClient(appConfig, destinationVault)

	secret, err := client.Logical().Delete(secretPath)
	if err != nil {
		return err
	}

	log.Infof("secret %s deleted", secret)

	return nil
}

// ListSecrets lists secrets located at the provided path
func (v *Client) ListSecrets(appConfig *config.AppConfig, destinationVault bool) {
	client = getClient(appConfig, destinationVault)

	secretsList, err := client.Logical().List(secretPath)
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range secretsList.Data {
		log.Printf("%v %v\n", k, v)
	}
}

// ListVaultMounts lists the mounts within the vault
func (v *Client) ListVaultMounts(appConfig *config.AppConfig, destinationVault bool) (vaultMounts map[string]*api.MountOutput, err error) {
	client = getClient(appConfig, destinationVault)

	mountsList, err := client.Sys().ListMounts()
	if err != nil {
		log.Fatal(err)
	}

	return mountsList, err
}

// DumpSecrets dumps all the secrets within the vault recursiviely
func (v *Client) DumpSecrets(appConfig *config.AppConfig, destinationVault bool) {
	client = getClient(appConfig, destinationVault)

	dumpNode(client, entryPoint)
}

// SyncSecret syncs a single secret from source to destination vault
func (v *Client) SyncSecret(appConfig *config.AppConfig, path string) error {
	log.Debugf("sync the secret %s", path)

	// get the secret from the source
	secret, err := v.ReadSecret(appConfig, path, false)
	if err != nil {
		log.Fatalf("failed to get secret %s from source vault", path, err)
	}
	if secret == nil {
		log.Fatalf("secret %s appears to be empty or does not exist in the source vault", path)
	}

	// WARNING: insecure
	log.Debug(secret, err)

	// get the secret from the destination, if it exists
	destSecret, err := v.ReadSecret(appConfig, path, true)
		if err != nil {
		log.Debugf("%s: destination secret likely doesn't exist", err)
	}
	// WARNING: insecure
	log.Debug(destSecret, err)

	// when the secret doesn't exist or the values are not the same
	// TODO: below assumes the existing secret is of kv2!
	if destSecret == nil || ! reflect.DeepEqual(secret.Data, destSecret.Data["data"]) {
		log.Debug("secret appears to need sync")
		err := writeSecret(appConfig.Destination.Client, path, secret.Data)
		if err != nil {
			log.Fatal("failed to write secret", err)
		}
		log.Debugf("secret written to %s", path)
		log.Infof("secret %s successfully sync'd", path)
	} else {
		log.Info("secret appears to be up to date, no sync required")
	}

	return nil
}

// SyncSecrets syncs all secrets from source to destination vault
func (v *Client) SyncSecrets(appConfig *config.AppConfig) {
	path := appConfig.Source.VaultEntrypoint
	log.Debugf("sync from entrypoint %s", path)
	syncNode(v, appConfig, path)
}
