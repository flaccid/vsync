package vault

import (
	log "github.com/Sirupsen/logrus"
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
	log.Debugf("write the secret: %s, %v", secret.Path, secret.Values)

	_, err := v.client.Logical().Write(secret.Path, secret.Values)
	if err != nil {
		return err
	}

	return nil
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
