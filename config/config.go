package config

import (
	"github.com/hashicorp/vault/api"
)

// AppConfig is the global application config
// which also includes the vault api config
type AppConfig struct {
	Credentials   string
	Filename      string
	LogLevel      string
	Vault         *api.Config
	VaultPassword string
	VaultToken    string
	VaultUsername string
}
