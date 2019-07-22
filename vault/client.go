package vault

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/flaccid/vsync/config"
	"github.com/hashicorp/vault/api"
)

// TODO: look at completing the remaining auth methods from
// https://github.com/cloudwatt/vault-sync/blob/master/pkg/vault/vault.go

func New(c *config.AppConfig) (*Client, error) {
	log.Debugf("create vault client to: %s", c.Vault.Address)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("unable to detect home directory", err)
	}

	if len(c.VaultToken) < 1 {
		data, err := ioutil.ReadFile(homeDir + "/.vault-token")
		if err != nil {
			log.Fatalf("file reading error", err)
		}
		c.VaultToken = string(data)
	}

	// uncomment to debug vault token (insecure!)
	// log.Debug("vault token: "+c.VaultToken)

	// step: get the client configuration
	config := api.DefaultConfig()
	config.Address = c.Vault.Address
	config.HttpClient = &http.Client{
		Timeout: time.Duration(15) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// step: get the client
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	// step: set the tocken for the client to use
	client.SetToken(c.VaultToken)

	return &Client{
		client: client,
	}, err
}
