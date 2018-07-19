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
	"path/filepath"
)

var flags = flag.NewFlagSet("Vaultsmith", flag.ExitOnError)
var documentPath string
var dry bool
var templateFile string
var vaultRole string
var logLevel string
var templateParams []string

func init() {
	flags.StringVar(
		// TODO: remove default value of "./example", could do bad things in production
		&documentPath, "document-path", "",
		"The root directory of the configuration. Can be a local directory, local gz tarball or http url to a gz tarball.",
	)
	flags.StringVar(
		&vaultRole, "role", "root", "The Vault role to authenticate as",
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
	flags.StringSliceVar(
		&templateParams, "template-params", []string{}, "Template parameters. Applies globally, but values in template-file take precedence. E.G.: service=foo,account=bar",
	)

	flags.Usage = func() {
		fmt.Printf("Usage of vaultsmith:\n")
		flags.PrintDefaults()
		fmt.Print("\nNotes:\n" +
			"• BE CAREFUL with this tool, it will faithfully apply whatever config you give it " +
			"without confirmation or warning! Use --dry until you are confident.\n" +
			"• Vault authentication is handled by environment variables (the same " +
			"ones as the Vault client, as vaultsmith uses the same code). So ensure VAULT_ADDR " +
			"and VAULT_TOKEN are set.\n" +
			"• If template-file is not specified, it is not mandatory for template.json to be present." +
			"• Specifying a parameter with --template-params allows only a single value. If you " +
			"need multiple values, please use a template-file." +
			"\n\n")
	}

	// Avoid parsing flags passed on running `go test`
	var args []string

	for _, s := range os.Args[1:] {
		if !strings.HasPrefix(s, "-test.") {
			args = append(args, s)
		}
	}

	err := flags.Parse(args)
	if err != nil {
		log.Fatal(err)
	}
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
	if documentPath == "" {
		log.Fatalln("Please specify --document-path")
	}
	// Only check if specified, otherwise no template file is OK
	if templateFile != "" {
		if _, err := os.Stat(templateFile); os.IsNotExist(err) {
			log.Fatalf("Specified template-file does not exist: %s", err)
		}
	}

	conf := config.VaultsmithConfig{
		DocumentPath:   documentPath,
		VaultRole:      vaultRole,
		TemplateFile:   templateFile,
		Dry:            dry,
		TemplateParams: templateParams,
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

func whichFileExists(filePath ...string) (file string) {
	for _, f := range filePath {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			log.WithFields(log.Fields{"file": f}).Debug("file exists")
			return f
		} else {
			log.WithFields(log.Fields{"file": f}).Debug("file does not exist")
		}
	}
	return ""
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

	docSet, err := document.GetSet(config.DocumentPath, workDir)
	if err != nil {
		return err
	}
	docSet.Get()
	defer docSet.CleanUp()

	// Determine if we have a template file
	config.TemplateFile = whichFileExists(
		templateFile,
		filepath.Join(docSet.Path(), "template.json"),
	)

	cw, err := internal.NewConfigWalker(c, config, docSet.Path())
	if err != nil {
		return err
	}
	return cw.Run()
}
