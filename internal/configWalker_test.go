package internal

import (
	"testing"
	"os"
	"time"
)

func TestWalkFile(t *testing.T) {
	cw := ConfigWalker{
		HandlerMap: map[string]PathHandler{
			"auth": MockPathHandler{},
			"aws": MockPathHandler{},
		},
	}
	f := &fakeFileInfo{}

	cw.walkFile("auth", f, nil)

}

// Mocks
type MockPathHandler struct {}

func (ph MockPathHandler) PutPoliciesFromDir(path string) error {
	return nil
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
