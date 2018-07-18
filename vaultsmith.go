package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"os"
	"strings"

	"github.com/starlingbank/vaultsmith/config"
	"github.com/starlingbank/vaultsmith/document"
	"github.com/starlingbank/vaultsmith/internal"
	"github.com/starlingbank/vaultsmith/vault"
	"io/ioutil"
	"net/url"
)

var flags = flag.NewFlagSet("Vaultsmith", flag.ExitOnError)
var documentPath string
var dry bool
var templateFile string
var vaultRole string
var logLevel string

func init() {
	flags.StringVar(
		// TODO: remove default value of "./example", could do bad things in production
		&documentPath, "document-path", "./example",
		"The root directory of the configuration. Can be a local directory, local gz tarball or http url to a gz tarball.",
	)
	flags.StringVar(
		&vaultRole, "role", "", "The Vault role to authenticate as",
	)
	flags.StringVar(
		&templateFile, "template-file", "", "JSON file containing template mappings. If not specified, vaultsmith will look for \"template.json\" in the base of the document path.",
	)
	flags.BoolVar(
		&dry, "dry", false, "Dry run; will read from but not write to vault",
	)
	flags.StringVar(
		&logLevel, "log-level", "info", fmt.Sprintf("Log level, valid values are %+v", log.AllLevels),
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
	ll, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetLevel(ll)

	if dry {
		log.Info("Dry mode enabled, no changes will be made")
	}
	conf := config.VaultsmithConfig{
		DocumentPath: documentPath,
		VaultRole:    vaultRole,
		TemplateFile: templateFile,
		Dry:          dry,
	}

	var client vault.Vault
	client, err = vault.NewVaultClient(conf.Dry)
	if err != nil {
		log.Fatal(err)
	}

	err = Run(client, conf)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	log.Debugf("Success")
}

// Return the appropriate document.Set for the given path
func getDocumentSet(path string, workDir string) (docSet document.Set, err error) {
	u, err := url.Parse(path)
	if err != nil {
		log.Error(err)
	}

	switch u.Scheme {
	case "http", "https":
		return &document.HttpTarball{
			WorkDir: workDir,
			Url:     u,
		}, nil
	case "", "file":
		// local filesystem, handled below
	default:
		// what is this?
		return nil, fmt.Errorf("unhandled scheme %q", u.Scheme)
	}

	// From here we are assuming path points to the local file system
	p, err := os.Stat(path)
	switch mode := p.Mode(); {
	case mode.IsDir():
		// Should be an directory of files
		return &document.LocalFiles{
			WorkDir:   workDir,
			Directory: path,
		}, nil
	case mode.IsRegular():
		// Should be an archive
		return &document.LocalTarball{
			WorkDir:     workDir,
			ArchivePath: path,
		}, nil
	default:
		return nil, fmt.Errorf("don't know what to do with mode %s", mode)
	}
}

func Run(c vault.Vault, config config.VaultsmithConfig) error {
	err := c.Authenticate(config.VaultRole)
	if err != nil {
		return fmt.Errorf("failed authenticating with Vault: %s", err)
	}

	workDir, err := ioutil.TempDir(os.TempDir(), "vaultsmith-")
	if err != nil {
		return fmt.Errorf("could not create temp directory: %s", err)
	}
	defer os.Remove(workDir)

	docSet, err := getDocumentSet(config.DocumentPath, workDir)
	if err != nil {
		return err
	}
	docSet.Get()
	defer docSet.CleanUp()

	cw, err := internal.NewConfigWalker(c, config, docSet.Path())
	if err != nil {
		return err
	}
	return cw.Run()
}
