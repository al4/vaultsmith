package document

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestLocalFiles_Get(t *testing.T) {
	l := LocalFiles{".", "."}
	if err := l.Get(); err != nil {
		log.Fatalf("Error running Get: %s", err)
	}
}

func TestLocalFiles_Path(t *testing.T) {
	exp := "."
	l := LocalFiles{Directory: "."}
	r := l.Path()
	if r != "." {
		log.Fatalf("Expected %q, got %s", r, exp)
	}
}

func TestLocalFiles_CleanUp(t *testing.T) {
	l := LocalFiles{Directory: "."}
	l.CleanUp()
}
