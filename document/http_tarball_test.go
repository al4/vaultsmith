package document

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

type TestHttpHandler struct {
	DummyData string
}

func (h *TestHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/gzip")
	fmt.Fprintln(w, h.DummyData)
}

func TestHttpTarball_archivePath(t *testing.T) {
	url, _ := url.Parse("https://example.com/test.tgz")
	p := HttpTarball{
		WorkDir: "/tmp/test",
		Url:     url,
	}
	exp := "/tmp/test/test.tgz"
	r := p.archivePath()
	if r != exp {
		log.Fatalf("Bad file path, expected %q, got %q", exp, r)
	}
}

func TestHttpTarball_Get(t *testing.T) {
	expected := "dummy data"
	ts := httptest.NewServer(&TestHttpHandler{
		DummyData: expected,
	})
	url, _ := url.Parse(ts.URL + "/test-archive.tgz")
	tmpDir, err := ioutil.TempDir(os.TempDir(), "fetcher-")
	if err != nil {
		log.Fatalf("Could not create tempdir: %s", err)
	}
	p := HttpTarball{
		WorkDir: tmpDir,
		Url:     url,
	}

	p.Get()
	defer p.CleanUp()

	if _, err := os.Stat(p.archivePath()); os.IsNotExist(err) {
		log.Fatalf("Expected file %s to exist", p.archivePath())
	}
	c, err := ioutil.ReadFile(p.archivePath())
	if err != nil {
		log.Fatalf("Error reading file %s", err)
	}

	// Something is adding a newline. It appears in the file, (so it must be present in the response
	// Body), but shouldn't be a problem in real-world use
	if string(c) != expected+"\n" {
		log.Fatalf("Expected file contents to be %q, got %q", expected, c)
	}
}

func TestHttpTarball_extract(t *testing.T) {
}
