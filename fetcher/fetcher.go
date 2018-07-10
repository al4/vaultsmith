package fetcher

import (
	"fmt"
	"os"
	"net/http"
	"io"
	"net/url"
	"log"
	"strings"
)

type Package struct {
	WorkDir	string
	Url		*url.URL
}

type Fetcher interface {
	Fetch() error
	Extract() error
	ExtractPath() string
	CleanUp()
}


// download tarball from Github
func (p *Package) Fetch() (err error){
	err = p.download()
	if err != nil {
		return fmt.Errorf("error downloading tarball %s", err)
	}
	return nil
}

func (p *Package) Extract() (path string, err error){

	return
}

// Return the path to the extracted files
// This does not ensure that the extract has actually been performed, so there is no guarantee the
// directory exists
func (p *Package) ExtractPath() (path string){
	s := strings.Split(
		strings.TrimRight(p.Url.Path, "/"),
		"/")

	dir := strings.TrimRight(p.WorkDir, string(os.PathSeparator))
	subdir := "extract-" + s[len(s) - 1]

	dirSlice := []string{dir, subdir}
	return strings.Join(dirSlice, string(os.PathSeparator))
}

func (p *Package) CleanUp() {
	log.Printf("Removing %s", p.WorkDir)
	err := os.RemoveAll(p.WorkDir)
	if err != nil {
		log.Println(err)
	}
	return
}

func (p *Package) download() error {
	out, err := os.Create(p.filePath())
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
	log.Printf("%v bytes written to %s", n, p.filePath())

	return nil
}

func (p *Package) filePath() (path string) {
	s := strings.Split(
		strings.TrimRight(p.Url.Path, "/"),
		"/")

	dir := strings.TrimRight(p.WorkDir, string(os.PathSeparator))
	file := s[len(s) - 1]

	ns := []string{dir, file}
	return strings.Join(ns, string(os.PathSeparator))
}

