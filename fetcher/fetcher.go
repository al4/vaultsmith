package fetcher

import (
	"fmt"
	"os"
	"net/http"
	"io"
	"net/url"
)

type Package struct {
	WorkDir	string
	Url		url.URL
}

type Fetcher interface {
	Get() error
	Extract() error
}


// download tarball from Github
func (p *Package) Get() {
	err := p.downloader("vaultsmith.tar")
	if err != nil {
		fmt.Errorf("error downloading tarball %s", err)
	}
}

func (p *Package) downloader(filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(p.Url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (p *Package) Extract() (path string, err error){

	return
}