package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/starlingbank/vaultsmith/internal"
	"github.com/starlingbank/vaultsmith/vault"
	"github.com/starlingbank/vaultsmith/config"
)

var flags = flag.NewFlagSet("Vaultsmith", flag.ExitOnError)
var documentPath	string
var vaultRole		string
var templateFile	string

func init() {
	flags.StringVar(
		// TODO: remove default value of "./example", could do bad things in production
		&documentPath, "document-path", "./example",
		"The root directory of the configuration. Can be a local directory or http url to a gzipped tarball.",
	)
	flags.StringVar(
		&vaultRole, "role", "", "The Vault role to authenticate as",
	)
	flags.StringVar(
		&templateFile, "templateFile", "example/template.json", "JSON file containing template mappings",
	)

	flags.Usage = func() {
		fmt.Printf("Usage of vaultsmith:\n")
		flags.PrintDefaults()
		fmt.Print("\nVault authentication is handled by environment variables (the same " +
			"ones as the Vault client, as vaultsmith uses the same code). So ensure VAULT_ADDR " +
			"and VAULT_TOKEN are set.\n\n")
	}

	// Avoid parsing flags passed on running `go test`
	var args []string
	for _, s := range os.Args[1:] {
		if !strings.HasPrefix(s, "-test.") {
			args = append(args, s)
		}
	}

	flags.Parse(args)
}

func main() {
	log.SetOutput(os.Stderr)

	conf := &config.VaultsmithConfig{
		DocumentPath: documentPath,
		VaultRole:    vaultRole,
		TemplateFile: templateFile,
	}

	var client *vault.Client
	client, err := vault.NewVaultClient()
	if err != nil {
		log.Fatal(err)
	}

	err = Run(client, conf)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	log.Println("Success")
}

func Run(c vault.Vault, config *config.VaultsmithConfig) error {
	err := c.Authenticate(config.VaultRole)
	if err != nil {
		return fmt.Errorf("failed authenticating with Vault: %s", err)
	}

	cw := internal.NewConfigWalker(c, config.DocumentPath)
	return cw.Run()
}
