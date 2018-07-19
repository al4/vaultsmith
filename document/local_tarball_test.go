package document

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalTarball_Get(t *testing.T) {
}

func TestLocalTarball_Path(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "test-vaultsmith-")
	if err != nil {
		log.Fatal(err)
	}
	lt := LocalTarball{
		WorkDir:     tmpDir,
		ArchivePath: filepath.Join(examplePath(), "/foo/test-foo-1.tgz"),
	}

	// mocking extraction
	td := filepath.Join(tmpDir, "test-foo-1.tgz-extract/foo")
	err = os.MkdirAll(td, 0755)
	if err != nil {
		log.Fatalf("Could not create directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	r := lt.Path()
	if r != td {
		log.Fatalf("Bad extract path, expected %q, got %q", td, r)
	}
}

func TestLocalTarball_CleanUp(t *testing.T) {
}

func TestLocalTarball_extract(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "test-vaultsmith-")
	if err != nil {
		log.Fatalf("Could not create temp dir: %s", err)
	}

	l := LocalTarball{
		WorkDir:     tmpDir,
		ArchivePath: filepath.Join(examplePath(), "/example.tar.gz"),
	}
	err = l.extract()
	defer l.CleanUp()
	if err != nil {
		log.Fatalf("Error calling extract: %s", err)
	}
	if _, err := os.Stat(l.Path()); os.IsNotExist(err) {
		log.Fatalf("Expected %s to exist", l.Path())
	}
}

func TestLocalTarball_extractPath(t *testing.T) {
	l := LocalTarball{
		WorkDir:     "/tmp/",
		ArchivePath: "/foo/test-foo-0.tgz",
	}
	exp := "/tmp/test-foo-0-extract"
	r := l.extractPath()
	if r != exp {
		log.Fatalf("Expected %q, got %q", exp, r)
	}
}
