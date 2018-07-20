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
		t.Errorf("Could not create directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	r, err := lt.Path()
	if err != nil {
		t.Errorf(err.Error())
	}
	if r != td {
		t.Errorf("Bad extract path, expected %q, got %q", td, r)
	}
}

func TestLocalTarball_CleanUp(t *testing.T) {
}

func TestLocalTarball_extract(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "test-vaultsmith-")
	if err != nil {
		t.Errorf("Could not create temp dir: %s", err)
	}

	l := LocalTarball{
		WorkDir:     tmpDir,
		ArchivePath: filepath.Join(examplePath(), "/example.tar.gz"),
	}
	err = l.extract()
	defer l.CleanUp()
	if err != nil {
		t.Errorf("Error calling extract: %s", err)
	}
	path, err := l.Path()
	if err != nil {
		t.Errorf(err.Error())
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected %s to exist", path)
	}
}

func TestLocalTarball_extractPath(t *testing.T) {
	l := LocalTarball{
		WorkDir:     "/tmp/",
		ArchivePath: "/foo/test-foo-0.tgz",
	}
	exp := "/tmp/test-foo-0.tgz-extract"
	r := l.extractPath()
	if r != exp {
		t.Errorf("Expected %q, got %q", exp, r)
	}
}
