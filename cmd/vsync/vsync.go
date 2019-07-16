package main

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/flaccid/vsync"
	"github.com/flaccid/vsync/config"
	"github.com/flaccid/vsync/vault"
	"github.com/hashicorp/vault/api"
	"github.com/urfave/cli"
)

var appConfig *config.AppConfig
var client *vault.Client

func beforeApp(c *cli.Context) error {
	level, err := log.ParseLevel(c.GlobalString("log-level"))
	if err != nil {
		log.Fatalf("unable to determine and set log level: %+v", err)
	}
	log.SetLevel(level)
	log.Debug("log level set to ", c.GlobalString("log-level"))

	// construct the application config here
	appConfig = &config.AppConfig{
		Vault: &api.Config{
			Address: c.String("vault-addr"),
		},
		VaultPassword: c.String("vault-password"),
		VaultToken:    c.String("vault-token"),
		VaultUsername: c.String("vault-username"),
		Credentials:   c.String("credentials"),
	}

	client, err = vault.New(appConfig)
	log.Debug(client)

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "vsync"
	app.Version = vsync.VERSION
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  vsync.AUTHOR,
			Email: vsync.EMAIL,
		},
	}
	app.Copyright = vsync.COPYRIGHT
	app.Usage = "hashicorp vault one-way secrets sync tool"
	app.UsageText = "vsync [global-options] [action]"
	app.ArgsUsage = "[vault action]"
	app.Action = start
	app.Before = beforeApp
	app.Commands = []cli.Command{
		cli.Command{
			Name:        "list-mounts",
			Aliases:     []string{"lm"},
			Usage:       "lists the vault mount points on the source vault server",
			UsageText:   "vsync --list-mounts",
			Description: "list the vault mounts",
			Action: func(c *cli.Context) error {
				client.ListVaultMounts()
				return nil
			},
		},
		cli.Command{
			Name:        "write-secret",
			Aliases:     []string{"as"},
			Usage:       "writes a single secret to the vault",
			UsageText:   "vsync --write-secret",
			Description: "write single secret",
			Action: func(c *cli.Context) error {
				secret := &vault.Secret{
					Path: "secret/foo",
					Values: map[string]interface{}{
						"value": "world",
						"foo":   "bar",
						"age":   "-1",
					},
				}
				client.WriteSecret(secret)
				return nil
			},
		},
		cli.Command{
			Name:        "dump-secrets",
			Aliases:     []string{"ds"},
			Usage:       "dumps all the secrets from the source vault server",
			UsageText:   "vsync --dump-secrets",
			Description: "dump em'",
			Action: func(c *cli.Context) error {
				client.DumpSecrets()
				return nil
			},
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "vault-addr,a",
			Usage:  "the url address of the vault service",
			Value:  "http://127.0.0.1:8200",
			EnvVar: "VAULT_ADDR",
		},
		cli.StringFlag{
			Name:   "vault-username,u",
			Usage:  "the vault username to use to authenticate to vault service",
			EnvVar: "VAULT_USERNAME",
		},
		cli.StringFlag{
			Name:   "vault-password,p",
			Usage:  "the vault password to use to authenticate to vault service",
			EnvVar: "VAULT_PASSWORD",
		},
		cli.StringFlag{
			Name:   "vault-token,t",
			Usage:  "a vault token used to authenticate to vault service",
			EnvVar: "VAULT_TOKEN",
		},
		cli.StringFlag{
			Name:   "credentials,c",
			Usage:  "the path to a file (json|yaml) containing the username and password for userpass authenticaion",
			EnvVar: "VAULT_CREDENTIALS",
		},
		cli.StringFlag{
			Name:  "log-level,l",
			Usage: "logging threshold level: debug|info|warn|error|fatal|panic",
			Value: "info",
		},
	}
	app.Run(os.Args)
}

func start(c *cli.Context) error {
	if len(c.Args().Get(0)) < 1 {
		log.Fatalf("please provide a command or use --help")
	}

	return nil
}
