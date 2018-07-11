package document

import (
	"strings"
	"path/filepath"
	"fmt"
	"os"
	"compress/gzip"
	"archive/tar"
	"io"
	"log"
)

// Implements document.Set
type LocalTarball struct {
	WorkDir string
	ArchivePath string
}

func (l *LocalTarball) Get() (err error) {
	l.extract()
	return nil
}

// Return the path to the extracted files. It does not guarantee that they exist.
func (l *LocalTarball) Path() (path string){
	return l.documentPath()
}

func (l *LocalTarball) CleanUp() {
	log.Printf("Removing %s", l.documentPath())
	err := os.RemoveAll(l.WorkDir)
	if err != nil {
		log.Println(err)
	}
	return
}

func (l *LocalTarball) extract() (err error){
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

	destDir := l.documentPath()

	for {
		hdr, err := tr.Next()
		if err == io.EOF { break }
		if err != nil {
			return fmt.Errorf("error reading tar archive %q: %s", l.ArchivePath, err)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:  // create dir
			dd := filepath.Join(destDir, hdr.Name)
			//log.Printf("Creating %q", dd)
			err := os.MkdirAll(dd, 0777)
			if err != nil {
				return fmt.Errorf("error creating directory %q: %s", dd, err)
			}
		case tar.TypeReg, tar.TypeRegA:
			df := filepath.Join(destDir, hdr.Name)
			log.Printf("Extracting %q", df)
			w, err := os.Create(df)
			if err != nil {
				return fmt.Errorf("error creating file %q: %s", df, err)
			}
			_, err = io.Copy(w, tr)
			if err != nil {
				return fmt.Errorf("error writing to file %q: %s", df, err)
			}
			w.Close()
		}
	}


	return
}

func (l *LocalTarball) documentPath() (path string) {
	_, file := filepath.Split(l.ArchivePath)

	name := strings.TrimSuffix(file, filepath.Ext(file))

	return filepath.Join(l.WorkDir, fmt.Sprintf("%s-extract", name))
}

