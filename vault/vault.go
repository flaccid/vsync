package vault

import (
	log "github.com/Sirupsen/logrus"
	"github.com/flaccid/vsync/config"
	"github.com/hashicorp/vault/api"
)

const (
	apiVersion = "v1"
)

// Client returns the underlining client
func (v *Client) Client() *api.Client {
	return v.client
}

// ReadSecret reads a single secret from the vault
func (v *Client) ReadSecret(path string) (*api.Secret, error) {
	secret, err := v.client.Logical().Read(path)

	return secret, err
}

// WriteSecret writes a single secret to the vault
func (v *Client) WriteSecret(secret *Secret) error {
	log.Debugf("write the secret to %s with %s", secret.Path, secret.Values)

	_, err := v.client.Logical().Write(secret.Path, secret.Values)
	if err != nil {
		return err
	}

	return nil
}

// writeSecret writes a single secret to the provided vault
// a private function that requires providing your vault api client
func writeSecret(v *api.Client, path string, secret *api.Secret) error {
	log.Debugf("write the secret to %s with %s", path, secret.Data)

	_, err := v.Logical().Write(path, secret.Data)
	if err != nil {
		return err
	}

	return nil
}

// ListSecrets lists secrets located at the provided path
func (v *Client) ListSecrets(path string) {
	secretsList, err := v.client.Logical().List(path)
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range secretsList.Data {
		log.Printf("%v %v\n", k, v)
	}
}

// ListVaultMounts lists the mounts within the vault
func (v *Client) ListVaultMounts() {
	mountsList, err := v.client.Sys().ListMounts()
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range mountsList {
		log.Printf("%v %v\n", k, v)
	}
}

// DumpSecrets dumps all the secrets within the vault recursiviely
func (v *Client) DumpSecrets(entryPoint string) {
	walkNode(v, entryPoint)
}

// SyncSecret syncs a single secret from source to destination vault
func (v *Client) SyncSecret(appConfig *config.AppConfig, path string) {
	// get the secret from the source
	secret, err := v.ReadSecret(path)
	// WARNING: insecure
	log.Debug(secret, err)

	// sync the secret to the destination vault
	err = writeSecret(appConfig.Destination.Client, path, secret)
	if err != nil {
		log.Fatal("failed to write secret", err)
	}
}

// SyncSecrets syncs all secrets from source to destination vault
func (v *Client) SyncSecrets(appConfig *config.AppConfig) {
	path := appConfig.VaultEntrypoint
	log.Debugf("sync %s", path)
	syncNode(v, appConfig, path)
}
