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
