package internal

import (
	"path/filepath"
	"os"
	"log"
	"fmt"
	"path"
	"strings"
)

type Walker interface {

}

type ConfigWalker struct {
	HandlerMap map[string]PathHandler
	Client     VaultsmithClient
	ConfigDir  string
}

func NewConfigWalker(client VaultsmithClient, configDir string) ConfigWalker {
	sysHandler, err := NewSysAuthHandler(client, filepath.Join(configDir, "sys"))
	if err != nil {
		log.Fatalf("Could not create syshandler: %s", err)
	}

	var handlerMap = map[string]PathHandler {
		"sys/auth": sysHandler,
	}
	log.Printf("%+v", handlerMap)


	return ConfigWalker{
		HandlerMap: handlerMap,
		Client: client,
		ConfigDir: path.Clean(configDir),
	}
}

func (cw ConfigWalker) Run() error {
	// file will be a dir here unless a trailing slash was added
	log.Printf("%+v", filepath.Join(cw.ConfigDir, "sys"))

	err := cw.walkConfigDir(cw.ConfigDir, cw.HandlerMap)
	if err != nil {
		return fmt.Errorf("error walking config dir %s: %s", cw.ConfigDir, err)
	}
	return nil
}

func (cw ConfigWalker) walkConfigDir(path string, handlerMap map[string]PathHandler) error {
	err := filepath.Walk(path, cw.walkFile)
	return err
}

// determine the handler and pass the root directory to it
func (cw ConfigWalker) walkFile(path string, f os.FileInfo, err error) error {
	//log.Printf("walking %s\n", path)
	relPath, err := filepath.Rel(cw.ConfigDir, path)
	if err != nil {
		return err
	}

	pathArray := strings.Split(relPath, string(os.PathSeparator))
	if ! f.IsDir() || len(pathArray) != 1 || pathArray[0] == "." {
		// we only want to operate on top-level directories, handler is responsible for walking
		return nil
	}
	log.Printf("path: %s, relPath: %s", path, relPath)

	handler, ok := cw.HandlerMap[pathArray[0]]
	if ! ok {
		log.Printf("no handler for path %s (file %s)", pathArray[0], relPath)
		return nil
	}
	log.Printf("Processing %s", path)
	return handler.PutPoliciesFromDir(path)
}
