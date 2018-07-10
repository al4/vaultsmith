package fetcher

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"fmt"
	"os"
	"net/url"
	"log"
	"io/ioutil"
)

type TestHttpHandler struct {
	DummyData string
}

func (h *TestHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/gzip")
	fmt.Fprintln(w, h.DummyData)
}

func TestPackage_filePath(t *testing.T) {
	url, _ := url.Parse("https://example.com/test.tgz")
	p := Package{
		WorkDir: "/tmp/test",
		Url: url,
	}
	exp := "/tmp/test/test.tgz"
	r := p.filePath()
	if r != exp {
		log.Fatalf("Bad file path, expected %q, got %q", exp, r)
	}
}

func TestPackage_ExtractPath(t *testing.T) {
	url, _ := url.Parse("https://example.com/test.tgz")
	p := Package{
		WorkDir: "/tmp/test",
		Url: url,
	}
	exp := "/tmp/test/extract-test.tgz"
	r := p.ExtractPath()
	if r != exp {
		log.Fatalf("Bad extract path, expected %q, got %q", exp, r)
	}
}

func TestPackage_Fetch(t *testing.T) {
	expected := "dummy data"
	ts := httptest.NewServer(&TestHttpHandler{
		DummyData: expected,
	})
	url, _ := url.Parse(ts.URL + "/test-archive.tgz")
	tmpDir, err := ioutil.TempDir(os.TempDir(), "fetcher-")
	if err != nil {
		log.Fatalf("Could not create tempdir: %s", err)
	}
	p := Package{
		WorkDir: tmpDir,
		Url: url,
	}

	p.Fetch()
	defer p.CleanUp()

	if _, err := os.Stat(p.filePath()); os.IsNotExist(err) {
		log.Fatalf("Expected file %s to exist", p.filePath())
	}
	c, err := ioutil.ReadFile(p.filePath())
	if err != nil {
		log.Fatalf("Error reading file %s", err)
	}

	// Something is adding a newline. It appears in the file, (so it must be present in the response
	// Body), but shouldn't be a problem in real-world use
	if string(c) != expected + "\n" {
		log.Fatalf("Expected file contents to be %q, got %q", expected, c)
	}
}

