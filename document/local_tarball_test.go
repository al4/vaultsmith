package document

import (
	"testing"
	"log"
)

func TestLocalTarball_Get(t *testing.T) {
}

func TestLocalTarball_Path(t *testing.T) {
}

func TestLocalTarball_CleanUp(t *testing.T) {
}

func TestLocalTarball_extract(t *testing.T) {
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

