package internal

import (
	log "github.com/sirupsen/logrus"
	"github.com/starlingbank/vaultsmith/path_handlers"
	"github.com/starlingbank/vaultsmith/vault"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestConfigHandlerWalkFile(t *testing.T) {
	cw := ConfigWalker{
		HandlerMap: map[string]path_handlers.PathHandler{
			"sys/auth": &path_handlers.Dummy{},
			"auth/aws": &path_handlers.Dummy{},
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
		HandlerMap: map[string]path_handlers.PathHandler{
			"parent": &path_handlers.Dummy{},
		},
	}

	hasParent := cw.hasParentHandler("parent/child")
	if !hasParent {
		log.Fatal("Got false for hasParent(\"parent/child\") call, should be true")
	}
}

// test hasParentHandler for the false case
func TestHasParentHandlerFalse(t *testing.T) {
	cw := ConfigWalker{
		HandlerMap: map[string]path_handlers.PathHandler{
			"parent": &path_handlers.Dummy{},
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
		HandlerMap: map[string]path_handlers.PathHandler{
			"parent/child": &path_handlers.Dummy{},
		},
	}

	hasParent := cw.hasParentHandler("parent/child")
	if hasParent {
		log.Fatal("Got true for hasParent(\"parent/child\") call, should be false (not parent of self).")
	}
}

// Should return true when there is no child handler
func TestConfigWalker_hasChildHandler(t *testing.T) {
	cw := ConfigWalker{
		HandlerMap: map[string]path_handlers.PathHandler{
			"base/parent/child": &path_handlers.Dummy{},
		},
	}

	hasChild := cw.hasChildHandler("base/parent")
	if !hasChild {
		log.Fatal("Got false for hasChildHandler call, should be true")
	}
}

// Should return false when there is no child handler
func TestConfigWalker_hasChildHandler_false(t *testing.T) {
	cw := ConfigWalker{
		HandlerMap: map[string]path_handlers.PathHandler{
			"foo": &path_handlers.Dummy{},
		},
	}

	hasChild := cw.hasChildHandler("base/parent")
	if hasChild {
		log.Fatal("Got true for hasChildHandler call, should be false")
	}
}

// Should not consider itself to be a child handler
func TestConfigWalker_hasChildHandler_self(t *testing.T) {
	cw := ConfigWalker{
		HandlerMap: map[string]path_handlers.PathHandler{
			"foo/bar": &path_handlers.Dummy{},
		},
	}

	hasChild := cw.hasChildHandler("foo/bar")
	if hasChild {
		log.Fatal("Got true for hasChildHandler call, should be false")
	}
}

func TestSortedPaths(t *testing.T) {
	fooH, err := path_handlers.NewDummyHandler(&vault.MockClient{}, "", 30)
	if err != nil {
		log.Fatal(err)
	}
	barH, err := path_handlers.NewDummyHandler(&vault.MockClient{}, "", 10)
	if err != nil {
		log.Fatal(err)
	}
	bozH, err := path_handlers.NewDummyHandler(&vault.MockClient{}, "", 20)
	if err != nil {
		log.Fatal(err)
	}

	cw := ConfigWalker{
		HandlerMap: map[string]path_handlers.PathHandler{
			"foo": fooH,
			"bar": barH,
			"boz": bozH,
		},
	}
	expected := []string{"bar", "boz", "foo"}
	r := cw.sortedPaths()
	if !reflect.DeepEqual(r, expected) {
		t.Errorf("Unexpected slice result (out of order?). Expected %+v; Got: %+v", expected, r)
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
