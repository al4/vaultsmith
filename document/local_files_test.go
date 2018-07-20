package document

import (
	"testing"
)

func TestLocalFiles_Get(t *testing.T) {
	l := LocalFiles{".", "."}
	if err := l.Get(); err != nil {
		t.Errorf("Error running Get: %s", err)
	}
}

func TestLocalFiles_Path(t *testing.T) {
	exp := "."
	l := LocalFiles{Directory: "."}
	r, err := l.Path()
	if err != nil {
		t.Error(err.Error())
	}
	if r != "." {
		t.Errorf("Expected %q, got %s", r, exp)
	}
}

func TestLocalFiles_CleanUp(t *testing.T) {
	l := LocalFiles{Directory: "."}
	l.CleanUp()
}
