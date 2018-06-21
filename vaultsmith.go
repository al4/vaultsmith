package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/starlingbank/vaultsmith/internal"
)

var flags = flag.NewFlagSet("Vaultsmith", flag.ExitOnError)
var configDir string
var vaultRole string

type VaultsmithConfig struct {
	configDir	string
	vaultRole	string
}

func NewVaultsmithConfig() (*VaultsmithConfig, error) {
	return &VaultsmithConfig{
		configDir: configDir,
		vaultRole: vaultRole,
	}, nil
}

func init() {
	flags.StringVar(
		// TODO: remove default value of "./example", could do bad things in prod
		&configDir, "configDir", "./example", "The root directory of the configuration",
	)
	flags.StringVar(
		&vaultRole, "role", "", "The Vault role to authenticate as",
	)

	flags.Usage = func() {
		fmt.Printf("Usage of vaultsmith:\n")
		flags.PrintDefaults()
		fmt.Println("\nVault authentication is handled by environment variables (the same " +
			"ones as the Vault client, as vaultsmith uses the same code). So ensure VAULT_ADDR " +
			"and VAULT_TOKEN are set.\n")
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

	config, err := NewVaultsmithConfig()
	if err != nil {
		log.Fatal(err)
	}

	var vaultClient *internal.VaultClient
	vaultClient, err = internal.NewVaultClient()
	if err != nil {
		log.Fatal(err)
	}

	err = Run(vaultClient, config)
	if err != nil {
		log.Fatal(err)
	}

}

func Run(c internal.VaultsmithClient, config *VaultsmithConfig) error {
	err := c.Authenticate(config.vaultRole)
	if err != nil {
		return fmt.Errorf("failed authenticating with Vault: %s", err)
	}

	//err = sysHandler.PutPoliciesFromDir("./example")
	//if err != nil {
	//	log.Fatal(err)
	//}
	cw := internal.NewConfigWalker(c, config.configDir)
	err = cw.Run()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Success")
	return nil
}
