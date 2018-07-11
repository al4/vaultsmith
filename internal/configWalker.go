package internal

import (
	"path/filepath"
	"os"
	"log"
	"fmt"
	"path"
	"strings"
	"github.com/starlingbank/vaultsmith/path_handlers"
	"github.com/starlingbank/vaultsmith/vault"
	"sort"
	"github.com/starlingbank/vaultsmith/config"
)

type Walker interface {

}

type ConfigWalker struct {
	HandlerMap map[string]path_handlers.PathHandler
	Client     vault.Vault
	ConfigDir  string
	Visited    map[string]bool
}

// Instantiates a configWalker and the required handlers
// TODO perhaps only instantiate if the path exists?
func NewConfigWalker(client vault.Vault, configDir string) ConfigWalker {
	sysAuthHandler, err := path_handlers.NewSysAuthHandler(client, filepath.Join(configDir, "sys", "auth"))
	if err != nil {
		log.Fatalf("Could not create sysAuthHandler: %s", err)
	}
	sysPolicyHandler, err := path_handlers.NewSysPolicyHandler(client, filepath.Join(configDir, "sys", "policy"))
	if err != nil {
		log.Fatalf("Could not create sysPolicyHandler: %s", err)
	}
	genericHandler, err := path_handlers.NewGenericHandler(client, config.VaultsmithConfig{ConfigDir: configDir}, filepath.Join(configDir, "auth"))
	if err != nil {
		log.Fatalf("Could not create genericHandler: %s", err)
	}

	var handlerMap = map[string]path_handlers.PathHandler {
		"auth": genericHandler,
		"sys/auth": sysAuthHandler,
		"sys/policy": sysPolicyHandler,
	}

	return ConfigWalker{
		HandlerMap: handlerMap,
		Client: client,
		ConfigDir: path.Clean(configDir),
		Visited: map[string]bool{},
	}
}

func (cw ConfigWalker) Run() error {
	// file will be a dir here unless a trailing slash was added
	log.Printf("Starting in directory %s", cw.ConfigDir)

	err := cw.walkConfigDir(cw.ConfigDir, cw.HandlerMap)
	if err != nil {
		return err
	}
	return nil
}

// Return a sorted slice of paths based on the Order() of its handler
func (cw ConfigWalker) sortedPaths() (paths []string) {
	for p := range cw.HandlerMap {
		paths = append(paths, p)
	}

	sort.Slice(paths, func(i, j int) bool {
		h1 := cw.HandlerMap[paths[i]]
		h2 := cw.HandlerMap[paths[j]]
		// zero (default) values always last
		if h1.Order() == 0 {
			return false
		}
		if h2.Order() == 0 {
			return true
		}
		return h1.Order() < h2.Order()
	})

	return paths
}

func (cw ConfigWalker) walkConfigDir(path string, handlerMap map[string]path_handlers.PathHandler) error {
	// Process according to <handler>.Order()
	paths := cw.sortedPaths()
	for _, v := range paths {
		handler := cw.HandlerMap[v]
		p := filepath.Join(path, v)
		err := handler.PutPoliciesFromDir(p)
		if err != nil {
			return err
		}
		cw.Visited[p] = true
	}

	// Process other directories with the genericHandler
	err := filepath.Walk(path, cw.walkFile)
	return err
}

// determine the handler and pass the root directory to it
func (cw ConfigWalker) walkFile(path string, f os.FileInfo, err error) error {
	if ! f.IsDir() {  // only want to operate on directories
		return nil
	}
	if visited, ok := cw.Visited[path]; ok && visited { // already been here
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

