package document

import (
	"fmt"
	"os"
	"net/http"
	"io"
	"log"
	"strings"
	"net/url"
)

// Implements document.Set
type HttpTarball struct {
	WorkDir	string
	Url		*url.URL
}

// download tarball from Github
func (p *HttpTarball) Get() (err error){
	err = p.download()
	if err != nil {
		return fmt.Errorf("error downloading tarball %s", err)
	}
	return nil
}

// Return the path to the extracted files
// This does not ensure that the extract has actually been performed, so there is no guarantee the
// directory exists
func (p *HttpTarball) Path() (path string){
	s := strings.Split(
		strings.TrimRight(p.Url.Path, "/"),
		"/")

	dir := strings.TrimRight(p.WorkDir, string(os.PathSeparator))
	subdir := "extract-" + s[len(s) - 1]

	dirSlice := []string{dir, subdir}
	return strings.Join(dirSlice, string(os.PathSeparator))
}

func (p *HttpTarball) CleanUp() {
	log.Printf("Removing %s", p.WorkDir)
	err := os.RemoveAll(p.WorkDir)
	if err != nil {
		log.Println(err)
	}
	return
}

func (p *HttpTarball) download() error {
	out, err := os.Create(p.archivePath())
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(p.Url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	log.Printf("%v bytes written to %s", n, p.archivePath())

	return nil
}

func (p *HttpTarball) extract() (err error){
	return
}


func (p *HttpTarball) archivePath() (path string) {
	s := strings.Split(
		strings.TrimRight(p.Url.Path, "/"),
		"/")

	dir := strings.TrimRight(p.WorkDir, string(os.PathSeparator))
	file := s[len(s) - 1]

	ns := []string{dir, file}
	return strings.Join(ns, string(os.PathSeparator))
}

