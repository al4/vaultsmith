package internal

import (
	"path/filepath"
	"os"
	"log"
	"fmt"
)

// A PathHandler takes a path and applies the policies within
type PathHandler interface {
	PutPoliciesFromDir(path string) error
}

type Walker interface {

}

type ConfigWalker struct {
	HandlerMap map[string]PathHandler
	Client     VaultsmithClient
	ConfigDir  string
}

func (cw *ConfigWalker) Run() error {
	// file will be a dir here unless a trailing slash was added
	log.Printf("%+v", filepath.Join(cw.ConfigDir, "sys"))

	err := cw.walkConfigDir(cw.ConfigDir, cw.HandlerMap)
	if err != nil {
		return fmt.Errorf("error walking config dir: %s", err)
	}
	return nil
}

func (cw *ConfigWalker) walkConfigDir(path string, handlerMap map[string]PathHandler) error {
	err := filepath.Walk(path, cw.walkFile)
	return err
}

func (cw *ConfigWalker) walkFile(path string, f os.FileInfo, err error) error {
	log.Printf("walking %s: %+v\n", path, f)

	return nil
}

