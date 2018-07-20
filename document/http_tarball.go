package document

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Implements document.Set
type HttpTarball struct {
	LocalTarball
	Url       *url.URL
	AuthToken string
}

// download tarball from Github
func (h *HttpTarball) Get() (err error) {
	downloadPath, err := h.download()
	if err != nil {
		return fmt.Errorf("error downloading tarball: %s", err)
	}

	h.LocalTarball.ArchivePath = downloadPath
	err = h.LocalTarball.extract()
	if err != nil {
		return fmt.Errorf("error extracting tarball: %s", err)
	}
	return nil
}

// Return the path to the extracted files. It does not guarantee that the path exists.
func (h *HttpTarball) Path() (path string, err error) {
	return h.LocalTarball.Path()
}

func (h *HttpTarball) CleanUp() {
	log.Infof("Removing %s", h.archivePath())
	err := os.RemoveAll(h.WorkDir)
	if err != nil {
		log.Error(err)
	}
	h.LocalTarball.CleanUp()
	return
}

func (h *HttpTarball) download() (path string, err error) {
	log.Infof("Downloading from %s to %s", h.Url.String(), h.archivePath())
	out, err := os.Create(h.archivePath())
	if err != nil {
		return "", err
	}
	defer out.Close()

	client := &http.Client{}
	req, err := http.NewRequest("GET", h.Url.String(), nil)
	if err != nil {
		return "", err
	}
	if h.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", h.AuthToken))
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code %v", res.StatusCode)
	}

	n, err := io.Copy(out, res.Body)
	if err != nil {
		return "", err
	}
	log.Infof("%v bytes written to %s", n, h.archivePath())

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
