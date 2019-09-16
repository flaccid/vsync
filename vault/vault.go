package vault

import (
	"errors"

	log "github.com/Sirupsen/logrus"
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
	if destinationVault {
		client = appConfig.Destination.Client
	} else {
		client = appConfig.Source.Client
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

	_, err := client.Logical().Write(secret.Path, secret.Values)
	if err != nil {
		return err
	}

	log.Infof("secret written to %s", secret.Path)

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
	if destinationVault {
		client = appConfig.Destination.Client
		secretPath = appConfig.Destination.VaultEntrypoint
	} else {
		client = appConfig.Source.Client
		secretPath = appConfig.Source.VaultEntrypoint
	}

	secretsList, err := client.Logical().List(secretPath)
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range secretsList.Data {
		log.Printf("%v %v\n", k, v)
	}
}

// ListVaultMounts lists the mounts within the vault
func (v *Client) ListVaultMounts(appConfig *config.AppConfig, destinationVault bool) {
	client = getClient(appConfig, destinationVault)

	mountsList, err := client.Sys().ListMounts()
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range mountsList {
		log.Printf("%v %v\n", k, v)
	}
}

// TODO: implement
// func (v *Client) GetMountProperties(appConfig *config.AppConfig, destinationVault bool, mount string) {
//
// }

// DumpSecrets dumps all the secrets within the vault recursiviely
func (v *Client) DumpSecrets(appConfig *config.AppConfig, destinationVault bool) {
	if destinationVault {
		client = appConfig.Destination.Client
		entryPoint = appConfig.Destination.VaultEntrypoint
	} else {
		client = appConfig.Source.Client
		entryPoint = appConfig.Source.VaultEntrypoint
	}

	dumpNode(client, entryPoint)
}

// SyncSecret syncs a single secret from source to destination vault
func (v *Client) SyncSecret(appConfig *config.AppConfig, path string) error {
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

	// sync the secret to the destination vault
	err = writeSecret(appConfig.Destination.Client, path, secret)
	if err != nil {
		log.Fatal("failed to write secret", err)
		return err
	}

	return nil
}

// SyncSecrets syncs all secrets from source to destination vault
func (v *Client) SyncSecrets(appConfig *config.AppConfig) {
	path := appConfig.Source.VaultEntrypoint
	log.Debugf("sync %s", path)
	syncNode(v, appConfig, path)
}
