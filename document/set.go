package document

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/starlingbank/vaultsmith/config"
	"net/url"
	"os"
)

// Retrieve the configuration files that we want to apply to Vault
type Set interface {
	Path() (string, error) // the path to the configuration documents. Should return nil if not present.
	Get() error            // fetch the configuration documents
	CleanUp()              // remove all temporary files
}

// Return the appropriate document.Set for the given path
func GetSet(workDir string, config config.VaultsmithConfig) (docSet Set, err error) {
	u, err := url.Parse(config.DocumentPath)
	if err != nil {
		log.Error(err)
	}

	switch u.Scheme {
	case "http", "https":
		return &HttpTarball{
			LocalTarball: LocalTarball{
				TarDir:  config.TarDir,
				WorkDir: workDir,
			},
			Url:       u,
			AuthToken: config.HttpAuthToken,
		}, nil
	case "", "file":
		// local filesystem, handled below
	default:
		// what is this?
		return nil, fmt.Errorf("unhandled scheme %q", u.Scheme)
	}

	// From here we are assuming path points to the local file system
	p, err := os.Stat(config.DocumentPath)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %s", config.DocumentPath, err)
	}
	switch mode := p.Mode(); {
	case mode.IsDir():
		// Should be an directory of files
		return &LocalFiles{
			WorkDir:   workDir,
			Directory: config.DocumentPath,
		}, nil
	case mode.IsRegular():
		// Should be an archive
		return &LocalTarball{
			WorkDir:     workDir,
			ArchivePath: config.DocumentPath,
			TarDir:      config.TarDir,
		}, nil
	default:
		return nil, fmt.Errorf("don't know what to do with mode %s", mode)
	}
}
