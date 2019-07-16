package vault

import "github.com/hashicorp/vault/api"

type Client struct {
	client *api.Client
}

type Secret struct {
	Path   string
	Values map[string]interface{}
}
