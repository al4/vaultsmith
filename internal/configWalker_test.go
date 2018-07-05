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

// test hasParentHandler for the true case
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

// test hasParentHandler for the false case
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

// test that we don't consider a directory to be a parent of itself
func TestHasParentHandlerSelf(t *testing.T) {
	cw := ConfigWalker{
		HandlerMap: map[string]handlers.PathHandler{
			"parent/child": &handlers.DummyHandler{},
		},
	}

	hasParent := cw.hasParentHandler("parent/child")
	if hasParent {
		log.Fatal("Got true for hasParent(\"parent/child\") call, should be false (not parent of self).")
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
