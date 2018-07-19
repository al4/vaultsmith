package document

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
)

// Retrieve the configuration files that we want to apply to Vault
type Set interface {
	Path() string // the path to the configuration documents. Should return nil if not present.
	Get() error   // fetch the configuration documents
	CleanUp()
}

// Return the appropriate document.Set for the given path
func GetSet(path string, workDir string) (docSet Set, err error) {
	u, err := url.Parse(path)
	if err != nil {
		log.Error(err)
	}

	switch u.Scheme {
	case "http", "https":
		return &HttpTarball{
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
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %s", path, err)
	}
	switch mode := p.Mode(); {
	case mode.IsDir():
		// Should be an directory of files
		return &LocalFiles{
			WorkDir:   workDir,
			Directory: path,
		}, nil
	case mode.IsRegular():
		// Should be an archive
		return &LocalTarball{
			WorkDir:     workDir,
			ArchivePath: path,
		}, nil
	default:
		return nil, fmt.Errorf("don't know what to do with mode %s", mode)
	}
}
