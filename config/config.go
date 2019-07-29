package config

import (
	"github.com/hashicorp/vault/api"
)

type VaultService struct {
	Client          *api.Client
	Vault           *api.Config
	VaultCredFile   string
	VaultPassword   string
	VaultToken      string
	VaultUsername   string
	VaultEntrypoint string
}

// AppConfig is the global application config
// which also includes the vault api config
type AppConfig struct {
	Source      *VaultService
	Destination *VaultService
	LogLevel    string
}
