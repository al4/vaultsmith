package document

import (
	"strings"
	"path/filepath"
	"fmt"
)

// Implements document.Set
type LocalTarball struct {
	WorkDir string
	ArchivePath string
}

func (l *LocalTarball) Get() (err error) {
	return nil
}

// Return the path to the extracted files. It does not guarantee that they exist.
func (l *LocalTarball) Path() (path string){
	return l.documentPath()
}

func (l *LocalTarball) CleanUp() {
	// NOOP for pre-existing local files!
	return
}

func (l *LocalTarball) extract() (err error){
	return
}

func (l *LocalTarball) documentPath() (path string) {
	_, file := filepath.Split(l.ArchivePath)

	name := strings.TrimSuffix(file, filepath.Ext(file))

	return filepath.Join(l.WorkDir, fmt.Sprintf("%s-extract", name))
}

