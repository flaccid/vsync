package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/flaccid/vsync"
	"github.com/flaccid/vsync/config"
	"github.com/flaccid/vsync/vault"
	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/pretty"
	"github.com/urfave/cli"
)

var (
	appConfig *config.AppConfig
	client    *vault.Client
	path      string
)

func beforeApp(c *cli.Context) error {
	// any validation do here
	level, err := log.ParseLevel(c.GlobalString("log-level"))
	if err != nil {
		log.Fatalf("unable to determine and set log level: %+v", err)
	}
	log.SetLevel(level)
	log.Debug("log level set to ", c.GlobalString("log-level"))

	// construct the application config here
	appConfig = &config.AppConfig{
		DryRun: c.Bool("dry"),
		Source: &config.VaultService{
			Vault: &api.Config{
				Address: c.String("vault-addr"),
			},
			VaultCredFile:   c.String("credentials-file"),
			VaultPassword:   c.String("vault-password"),
			VaultToken:      c.String("vault-token"),
			VaultUsername:   c.String("vault-username"),
			VaultEntrypoint: c.String("entrypoint"),
		},
		Destination: &config.VaultService{
			Vault: &api.Config{
				Address: c.String("destination-vault-addr"),
			},
			VaultEntrypoint: c.String("entrypoint"),
			VaultPassword:   c.String("destination-vault-password"),
			VaultToken:      c.String("destination-vault-token"),
			VaultUsername:   c.String("destination-vault-username"),
		},
	}
	log.Debug(spew.Sdump(appConfig))

	appConfig.Source.Client, err = vault.New(appConfig)
	if err != nil {
		log.Fatalf("error creating source client: %+v", err)
	}
	log.Debug("source client", appConfig.Source.Client)

	if len(c.String("destination-vault-addr")) > 0 {
		appConfig.Destination.Client, err = vault.NewDest(appConfig)
		if err != nil {
			log.Fatalf("error creating destination client: %+v", err)
		}
		log.Debug(appConfig.Destination.Client)
	}

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
			Name:        "list",
			Aliases:     []string{"ls"},
			Usage:       "lists secrets in the entrypoint path",
			UsageText:   "vsync list",
			Description: "list secrets",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "destination-vault, ds",
					Usage: "peforms the operation on the destination vault"},
			},
			Action: func(c *cli.Context) error {
				client.ListSecrets(appConfig, c.Bool("destination-vault"))
				return nil
			},
		},
		cli.Command{
			Name:        "read-secret",
			Aliases:     []string{"rs"},
			Usage:       "reads a single secret from the source vault",
			UsageText:   "vsync read-secret",
			Description: "read single secret",
			ArgsUsage:   "[secret path]",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "destination-vault, ds",
					Usage: "peforms the operation on the destination vault"},
			},
			Action: func(c *cli.Context) error {
				path := c.Args().First()
				if len(path) < 1 {
					log.Fatal("please provide a secret path to read")
				}
				log.Debugf("read secret %s", path)

				secret, err := client.ReadSecret(appConfig, path, c.Bool("destination-vault"))
				if err != nil {
					log.Fatal(err)
				}

				log.Debugf("%s: %s", path, secret.Data)
				j, err := json.MarshalIndent(secret, "", "    ")
				if err != nil {
					log.Fatalf("error marshalling json: ", err.Error())
				}
				fmt.Printf("%s\n", string(j))

				return nil
			},
		},
		cli.Command{
			Name:        "write-secret",
			Aliases:     []string{"ws"},
			Usage:       "writes a single or multiple kv secret(s) to the vault",
			UsageText:   "vsync write-secret [path] [key=value,[[key=value]]]",
			Description: "write key/value secret(s)",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "destination-vault, ds",
					Usage: "peforms the operation on the destination vault"},
			},
			Action: func(c *cli.Context) error {
				// NOTE: this overwrites, it does not merge the keys of a secret

				if len(c.Args().First()) < 1 {
					log.Fatal("please provide a secret path to write")
				}
				if len(c.Args()) < 2 {
					log.Fatal("please provide a key=value for the secret")
				}

				// local variable for the secret values
				values := map[string]interface{}{}

				// initiate the secret without any values
				secret := &vault.Secret{
					Path: c.Args().First(),
				}

				// split out the cmdline secrets into pairs
				pairs := strings.Split(c.Args()[1], ",")
				log.Debugf("pairs: ", pairs)

				// iterate through the pairs and assign to the local var
				for _, pair := range pairs {
					kv := strings.Split(pair, "=")
					log.Debugf("pair split: %s, %s", kv[0], kv[1])
					values[kv[0]] = kv[1]
				}

				// assign the values in the secret from the local var
				secret.Values = values

				// finally, write the entire secret
				err := client.WriteSecret(appConfig, secret, c.Bool("destination-vault"))
				if err != nil {
					log.Fatal(err)
				}

				return nil
			},
		},
		cli.Command{
			Name:        "sync-secret",
			Aliases:     []string{"ss"},
			Usage:       "syncs a single secret from source to destination vault",
			UsageText:   "vsync sync-secret [path]",
			Description: "sync a single secret",
			Action: func(c *cli.Context) error {
				if len(c.Args().First()) < 1 {
					log.Fatal("please provide a secret path to sync")
				}
				path = c.Args().First()
				err := client.SyncSecret(appConfig, path)
				if err != nil {
					log.Fatal(err)
				}
				return nil
			},
		},
		cli.Command{
			Name:        "sync-secrets",
			Aliases:     []string{"s"},
			Usage:       "syncs all secrets to the destination vault",
			UsageText:   "vsync sync-secrets",
			Description: "sync all secrets",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "remove-orphans, ro",
					Usage: "removes orphans in the destination vault after sync"},
			},
			Action: func(c *cli.Context) error {
				client.SyncSecrets(appConfig)
				if c.Bool("remove-orphans") {
					log.Info("remove orphans in destination vault")
					log.Info("fetching all secrets in destination vault, please wait...")
					orphansRemoved, err := client.RemoveOrphans(appConfig, appConfig.Destination.VaultEntrypoint)
					if err != nil {
						log.Fatal(err)
					}
					log.Infof("%v orphans successfully removed", len(orphansRemoved))
				}
				return nil
			},
		},
		cli.Command{
			Name:        "dump-secrets",
			Aliases:     []string{"ds"},
			Usage:       "dumps all the secrets from the source vault server",
			UsageText:   "vsync dump-secrets",
			Description: "dump em'",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "destination-vault, ds",
					Usage: "peforms the operation on the destination vault"},
			},
			Action: func(c *cli.Context) error {
				client.DumpSecrets(appConfig, c.Bool("destination-vault"))
				return nil
			},
		},
		cli.Command{
			Name:        "remove-orphans",
			Aliases:     []string{"ro"},
			Usage:       "removes orphaned secret paths in the destination vault",
			UsageText:   "vsync remove-orphans",
			Description: "remove orphaned secret paths",
			Action: func(c *cli.Context) error {
				if len(appConfig.Destination.Vault.Address) < 1 {
					log.Fatal("please provide destination vault parameters")
				}
				log.Info("fetching all secrets in destination vault, please wait...")
				orphansRemoved, err := client.RemoveOrphans(appConfig, appConfig.Destination.VaultEntrypoint)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("%v orphans successfully removed", len(orphansRemoved))
				return nil
			},
		},
		cli.Command{
			Name:        "list-mounts",
			Aliases:     []string{"lm"},
			Usage:       "lists the vault mount points on the source vault server",
			UsageText:   "vsync list-mounts",
			Description: "list the vault mounts",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "destination-vault, ds",
					Usage: "peforms the operation on the destination vault"},
			},
			Action: func(c *cli.Context) error {
				vaultMounts, err := client.ListVaultMounts(appConfig, c.Bool("destination-vault"))
				if err != nil {
					log.Fatalf(err.Error())
				}
				j, err := vault.ToJson(vaultMounts)
				if err != nil {
					log.Fatalf(err.Error())
				}
				fmt.Println(string(j))
				return nil
			},
		},
		cli.Command{
			Name:        "health",
			Aliases:     []string{"hc"},
			Usage:       "perform a health check",
			UsageText:   "vsync health",
			Description: "performs a health check on the vault server",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "destination-vault, ds",
					Usage: "peform operation on the destination vault server"},
			},
			Action: func(c *cli.Context) error {
				health, err := client.HealthCheck(appConfig, c.Bool("destination-vault"))
				if err != nil {
					log.Fatalf("error getting health status: %s", err)
				}
				fmt.Println(health)
				return nil
			},
		},
		cli.Command{
			Name:        "request",
			Aliases:     []string{"req"},
			Usage:       "make a raw arbitrary request",
			UsageText:   "vsync request [type] [uri]",
			Description: "make a raw arbitrary request",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "destination-vault, ds",
					Usage: "peforms the operation on the destination vault"},
				cli.BoolFlag{Name: "api-version",
					Usage: "api version for the request"},
			},
			Action: func(c *cli.Context) error {
				if len(c.Args().Get(0)) < 2 {
					log.Fatalf("please provide the request type, e.g. GET")
				}
				if len(c.Args().Get(1)) < 2 {
					log.Fatalf("please provide the request uri, e.g. /sys/health")
				}
				log.Infof("request %s", c.Args())
				resp, err := client.Request(appConfig, c.Bool("destination-vault"), c.Args().Get(0), c.Args().Get(1), "")
				if err != nil {
					log.Fatalf("error making http request: %s", err)
				}
				log.Debug(resp, err)
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatalf("error reading http response body: %s", err)
				}
				fmt.Println(string(pretty.Color(pretty.Pretty(body), nil)))
				return nil
			},
		},
		cli.Command{
			Name:        "show-config",
			Aliases:     []string{"sc"},
			Usage:       "show a summary of the config",
			UsageText:   "vsync show-config",
			Description: "print config",
			Action: func(c *cli.Context) error {
				// TODO: pretty print out config and mask secrets
				// currently using spew and is insecure
				log.Info(spew.Sdump(appConfig))
				return nil
			},
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "vault-addr,a",
			Usage:  "url address of the source vault service",
			Value:  "http://127.0.0.1:8200",
			EnvVar: "VAULT_ADDR",
		},
		cli.StringFlag{
			Name:   "vault-token,t",
			Usage:  "vault token used to authenticate to source vault service",
			EnvVar: "VAULT_TOKEN",
		},
		cli.StringFlag{
			Name:   "vault-username,u",
			Usage:  "vault username to use to authenticate to source vault service",
			EnvVar: "VAULT_USERNAME",
		},
		cli.StringFlag{
			Name:   "vault-password,p",
			Usage:  "vault password to use to authenticate to source vault service",
			EnvVar: "VAULT_PASSWORD",
		},
		cli.StringFlag{
			Name:   "credentials-file,c",
			Usage:  "path to a file (json|yaml) containing the username and password for userpass authentication",
			EnvVar: "VAULT_CREDENTIALS",
		},
		cli.StringFlag{
			Name:   "entrypoint,e",
			Usage:  "vault entry point path",
			EnvVar: "VAULT_PREFIX",
			Value:  "/secret",
		},
		cli.StringFlag{
			Name:   "destination-vault-addr",
			Usage:  "destination vault url",
			EnvVar: "DESTINATION_VAULT_ADDR",
		},
		cli.StringFlag{
			Name:   "destination-vault-token",
			Usage:  "destination vault token",
			EnvVar: "DESTINATION_VAULT_TOKEN",
		},
		cli.StringFlag{
			Name:   "destination-vault-username",
			Usage:  "destination vault username",
			EnvVar: "DESTINATION_VAULT_USERNAME",
		},
		cli.StringFlag{
			Name:   "destination-vault-password",
			Usage:  "destination vault password",
			EnvVar: "DESTINATION_VAULT_PASSWORD",
		},
		cli.BoolFlag{
			Name:   "dry",
			Usage:  "dry run",
			EnvVar: "VSYNC_DRY_RUN",
		},
		cli.StringFlag{
			Name:   "log-level,l",
			Usage:  "logging threshold level: debug|info|warn|error|fatal|panic",
			EnvVar: "VSYNC_LOG_LEVEL",
			Value:  "info",
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
