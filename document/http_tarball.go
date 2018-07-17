package document

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Implements document.Set
type HttpTarball struct {
	LocalTarball
	WorkDir string
	Url     *url.URL
}

// download tarball from Github
func (h *HttpTarball) Get() (err error) {
	downloadPath, err := h.download()
	if err != nil {
		return fmt.Errorf("error downloading tarball %s", err)
	}

	h.LocalTarball = LocalTarball{
		WorkDir:     h.WorkDir,
		ArchivePath: downloadPath,
	}
	err = h.LocalTarball.extract()
	if err != nil {
		return fmt.Errorf("error extracting tarball %s", err)
	}
	return nil
}

// Return the path to the extracted files. It does not guarantee that the path exists.
func (h *HttpTarball) Path() (path string) {
	return h.LocalTarball.documentPath()
}

func (h *HttpTarball) CleanUp() {
	log.Printf("Removing %s", h.archivePath())
	err := os.RemoveAll(h.WorkDir)
	if err != nil {
		log.Println(err)
	}
	h.LocalTarball.CleanUp()
	return
}

func (h *HttpTarball) download() (path string, err error) {
	log.Printf("Downloading from %s to %s", h.Url.String(), h.archivePath())
	out, err := os.Create(h.archivePath())
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(h.Url.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	log.Printf("%v bytes written to %s", n, h.archivePath())

	return out.Name(), nil
}

func (h *HttpTarball) archivePath() (path string) {
	s := strings.Split(
		strings.TrimRight(h.Url.Path, "/"),
		"/")

	dir := strings.TrimRight(h.WorkDir, string(os.PathSeparator))
	file := s[len(s)-1]

	ns := []string{dir, file}
	return strings.Join(ns, string(os.PathSeparator))
}
