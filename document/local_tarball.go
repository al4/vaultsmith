package document

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Implements document.Set
type LocalTarball struct {
	WorkDir     string
	ArchivePath string
}

func (l *LocalTarball) Get() (err error) {
	return l.extract()
}

// Return the path to the extracted files. It does not guarantee that they exist.
func (l *LocalTarball) Path() (path string, err error) {
	// Most tarballs, including github tarballs, will contain a single directory with the archive
	// contents
	// TODO should probably have an option for this behaviour; what if a user only has one config dir?
	entries, err := ioutil.ReadDir(l.extractPath())
	if err != nil {
		return "", err
	}
	if len(entries) == 1 && entries[0].Name() != "sys" && entries[0].IsDir() {
		// Probably a single dir, use it instead
		return filepath.Join(l.extractPath(), entries[0].Name()), err
	} else if len(entries) > 1 {
		// More than one entry suggests we already have the correct path
		return l.extractPath(), err
	} else {
		log.Warnf("Empty directory %q", l.extractPath())
		return "", err
	}
}

func (l *LocalTarball) CleanUp() {
	log.Infof("Removing %s", l.extractPath())
	err := os.RemoveAll(l.WorkDir)
	if err != nil {
		log.Error(err)
	}
	return
}

func (l *LocalTarball) extract() (err error) {
	log.Debugf("Extracting %s", l.ArchivePath)
	f, err := os.Open(l.ArchivePath)
	if err != nil {
		return fmt.Errorf("could not open file %q: %s", l.ArchivePath, err)
	}
	r, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("could not create gzip reader for %q: %s", l.ArchivePath, err)
	}
	tr := tar.NewReader(r)
	if err != nil {
		return fmt.Errorf("could not create tar reader for %q: %s", l.ArchivePath, err)
	}

	destDir := l.extractPath()

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar archive %q: %s", l.ArchivePath, err)
		}
		switch hdr.Typeflag {
		case tar.TypeDir: // create dir
			dd := filepath.Join(destDir, hdr.Name)
			log.Debugf("Creating %q", dd)
			err := os.MkdirAll(dd, 0777)
			if err != nil {
				return fmt.Errorf("error creating directory %q: %s", dd, err)
			}
		case tar.TypeReg, tar.TypeRegA:
			df := filepath.Join(destDir, hdr.Name)
			log.Infof("Extracting %q", df)
			w, err := os.Create(df)
			if err != nil {
				return fmt.Errorf("error creating file %q: %s", df, err)
			}
			_, err = io.Copy(w, tr)
			if err != nil {
				return fmt.Errorf("error writing to file %q: %s", df, err)
			}
			w.Close()
		default:
			log.Debugf("Unhandled tar type: %+v", hdr)
		}
	}

	return
}

func (l *LocalTarball) extractPath() (path string) {
	_, file := filepath.Split(l.ArchivePath)

	return filepath.Join(l.WorkDir, fmt.Sprintf("%s-extract", file))
}
