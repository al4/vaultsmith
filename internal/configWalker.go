package internal

import (
	"path/filepath"
	"os"
	"log"
	"fmt"
	"path"
	"strings"
	"github.com/starlingbank/vaultsmith/handlers"
	"github.com/starlingbank/vaultsmith/vaultClient"
)

type Walker interface {

}

type ConfigWalker struct {
	HandlerMap map[string]handlers.PathHandler
	Client     vaultClient.VaultsmithClient
	ConfigDir  string
}

func NewConfigWalker(client vaultClient.VaultsmithClient, configDir string) ConfigWalker {
	sysHandler, err := handlers.NewSysAuthHandler(client, filepath.Join(configDir, "sys"))
	if err != nil {
		log.Fatalf("Could not create syshandler: %s", err)
	}

	var handlerMap = map[string]handlers.PathHandler {
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

func (cw ConfigWalker) walkConfigDir(path string, handlerMap map[string]handlers.PathHandler) error {
	err := filepath.Walk(path, cw.walkFile)
	return err
}

// determine the handler and pass the root directory to it
func (cw ConfigWalker) walkFile(path string, f os.FileInfo, err error) error {
	if ! f.IsDir() {  // only want to operate on directories
		return nil
	}
	relPath, err := filepath.Rel(cw.ConfigDir, path)
	if err != nil {
		return fmt.Errorf("could not determine relative path of %s to %s: %s", path, cw.ConfigDir, err)
	}

	pathArray := strings.Split(relPath, string(os.PathSeparator))
	if pathArray[0] == "." { // just to avoid a "no handler for path ." in log
		log.Printf("Skipping %s", relPath)
		return nil
	}
	//log.Printf("path: %s, relPath: %s", path, relPath)

	// Is there a handler for a higher level path? If so, we assume that it handles all child
	// directories and thus we should not process these directories separately.
	if hasParentHandler(relPath) {
		return nil
	}

	handler, ok := cw.HandlerMap[relPath]
	if ! ok {
		log.Printf("no handler for path %s (file %s)", relPath, relPath)
		return nil
	}
	log.Printf("Processing %s", path)
	return handler.PutPoliciesFromDir(path)
}


func hasParentHandler(string) bool {
	return false
}

