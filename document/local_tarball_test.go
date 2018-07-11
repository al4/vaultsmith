package document

import (
	"testing"
	"log"
	"io/ioutil"
	"os"
	"path/filepath"
)

func TestLocalTarball_Get(t *testing.T) {
}

func TestLocalTarball_Path(t *testing.T) {
}

func TestLocalTarball_CleanUp(t *testing.T) {
}

func TestLocalTarball_extract(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "test-vaultsmith-")
	if err != nil {
		log.Fatalf("Could not create temp dir: %s", err)
	}

	l := LocalTarball{
		WorkDir: tmpDir,
		ArchivePath: filepath.Join(examplePath(), "/example.tar.gz"),
	}
	log.Println(l.ArchivePath)
	log.Println(l.documentPath())
	err = l.extract()
	defer l.CleanUp()
	if err != nil {
		log.Fatalf("Error calling extract: %s", err)
	}
	if _, err := os.Stat(l.Path()); os.IsNotExist(err) {
		log.Fatalf("Expected %s to exist", l.Path())
	}
}

func TestLocalTarball_documentPath(t *testing.T) {
	l := LocalTarball{
		WorkDir: "/tmp/",
		ArchivePath: "/foo/test-foo-0.tgz",
	}
	exp := "/tmp/test-foo-0-extract"
	r := l.documentPath()
	if r !=  exp {
		log.Fatalf("Expected %q, got %q", exp, r)
	}
}

