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

// Instantiates a configWalker and the required handlers
// TODO perhaps only instantiate if the path exists?
func NewConfigWalker(client vaultClient.VaultsmithClient, configDir string) ConfigWalker {
	sysAuthHandler, err := handlers.NewSysAuthHandler(client, filepath.Join(configDir, "sys", "auth"))
	if err != nil {
		log.Fatalf("Could not create sysAuthHandler: %s", err)
	}
	sysPolicyHandler, err := handlers.NewSysPolicyHandler(client, filepath.Join(configDir, "sys", "policy"))
	if err != nil {
		log.Fatalf("Could not create sysPolicyHandler: %s", err)
	}

	var handlerMap = map[string]handlers.PathHandler {
		"sys/auth": sysAuthHandler,
		"sys/policy": sysPolicyHandler,
	}

	return ConfigWalker{
		HandlerMap: handlerMap,
		Client: client,
		ConfigDir: path.Clean(configDir),
	}
}

func (cw ConfigWalker) Run() error {
	// file will be a dir here unless a trailing slash was added
	log.Printf("Starting in directory %s", filepath.Join(cw.ConfigDir, "sys"))

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
		return nil
	}
	//log.Printf("path: %s, relPath: %s", path, relPath)

	// Is there a handler for a higher level path? If so, we assume that it handles all child
	// directories and thus we should not process these directories separately.
	if cw.hasParentHandler(relPath) {
		log.Printf("Skipping %s, handled by parent", relPath)
		return nil
	}

	handler, ok := cw.HandlerMap[relPath]
	if ! ok {
		log.Printf("No handler for path %s", relPath)
		return nil
	}
	log.Printf("Processing %s", path)
	return handler.PutPoliciesFromDir(path)
}

// Determine whether this directory is already covered by a parent handler
func (cw ConfigWalker) hasParentHandler(path string) bool {
	pathArr := strings.Split(path, string(os.PathSeparator))
	for i := 0; i < len(pathArr) - 1; i++ {
		s := strings.Join(pathArr[:i + 1], string(os.PathSeparator))
		if s == path { // should be covered by -1 in for condition above, so just for extra safety
			continue
		}
		if _, ok := cw.HandlerMap[s]; ok {
			return true
		}
	}
	return false
}

