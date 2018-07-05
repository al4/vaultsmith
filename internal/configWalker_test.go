package internal

import (
	"testing"
	"os"
	"time"
	"log"
	"github.com/starlingbank/vaultsmith/handlers"
)

func TestConfigHandlerWalkFile(t *testing.T) {
	cw := ConfigWalker{
		HandlerMap: map[string]handlers.PathHandler{
			"sys/auth": &handlers.DummyHandler{},
			"auth/aws":  &handlers.DummyHandler{},
		},
	}
	f := &fakeFileInfo{}

	e := cw.walkFile("auth", f, nil)
	if e != nil {
		log.Fatal(e)
	}
}

func TestHasParentHandlerTrue(t *testing.T) {
	cw := ConfigWalker{
		HandlerMap: map[string]handlers.PathHandler{
			"parent": &handlers.DummyHandler{},
		},
	}

	hasParent := cw.hasParentHandler("parent/child")
	if ! hasParent {
		log.Fatal("Got false for hasParent(\"parent/child\") call, should be true")
	}
}

func TestHasParentHandlerFalse(t *testing.T) {
	cw := ConfigWalker{
		HandlerMap: map[string]handlers.PathHandler{
			"parent": &handlers.DummyHandler{},
		},
	}

	hasParent := cw.hasParentHandler("notparent/child")
	if hasParent {
		log.Fatal("Got true for hasParent(\"notparent/child\") call, should be false")
	}
}

type fakeFileInfo struct {
	dir      bool
	basename string
	modtime  time.Time
	ents     []*fakeFileInfo
	contents string
	err      error
}

func (f *fakeFileInfo) Name() string       { return f.basename }
func (f *fakeFileInfo) Sys() interface{}   { return nil }
func (f *fakeFileInfo) ModTime() time.Time { return f.modtime }
func (f *fakeFileInfo) IsDir() bool        { return f.dir }
func (f *fakeFileInfo) Size() int64        { return int64(len(f.contents)) }
func (f *fakeFileInfo) Mode() os.FileMode {
	if f.dir {
		return 0755 | os.ModeDir
	}
	return 0644
}
