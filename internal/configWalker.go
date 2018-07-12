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


// The ConfigWalker assumes it is in the root of the vault configuration to apply. As an example, it
// would directly contain "sys" and "auth".

type Walker interface {

}

type ConfigWalker struct {
	HandlerMap map[string]path_handlers.PathHandler
	Client     vault.Vault
	ConfigDir  string
	Visited    map[string]bool
}

// Instantiates a configWalker and the required handlers
// TODO this mixes configuration and code, could be declared in a better way
func NewConfigWalker(client vault.Vault, config config.VaultsmithConfig, docPath string) ConfigWalker {
	// Map configuration directories to specific path handlers
	var handlerMap = map[string]path_handlers.PathHandler{}

	// Instantiate our path handlers
	// We handle any unknown directories with this one
	genericHandler, err := path_handlers.NewGenericHandler(
		client,
		path_handlers.PathHandlerConfig{
			DocumentPath: docPath,
			MappingFile: config.TemplateFile,
		})
	if err != nil {
		log.Fatalf("Could not create genericHandler: %s", err)
	}
	handlerMap["*"] = genericHandler

	// sys directories should never be generic, so apply a dummy at the top level
	nullHandler, err := path_handlers.NewDummyHandler(client, "", 0)
	if err != nil {
		log.Fatalf("Error instantiating null handler: %s", err)
	}
	handlerMap["sys"] = nullHandler

	// The two sys path handlers
	sysAuthDir := filepath.Join(docPath, "sys", "auth")
	if f, err := os.Stat(sysAuthDir); ! os.IsNotExist(err) {
		if f.Mode().IsDir() {
			sysAuthHandler, err := path_handlers.NewSysAuthHandler(
				client,
				path_handlers.PathHandlerConfig{
					DocumentPath: docPath,
					Order: 10,
				})
			if err != nil {
				log.Fatalf("Could not create sysAuthHandler: %s", err)
			}
			handlerMap["sys/auth"] = sysAuthHandler
		}
	}

	sysPolicyDir := filepath.Join(docPath, "sys", "policy")
	if f, err := os.Stat(sysPolicyDir); ! os.IsNotExist(err) {
		if f.Mode().IsDir() {
			sysPolicyHandler, err := path_handlers.NewSysPolicyHandler(
				client,
				path_handlers.PathHandlerConfig{
					DocumentPath: docPath,
					Order: 20,
				})
			if err != nil {
				log.Fatalf("Could not create sysPolicyHandler: %s", err)
			}
			handlerMap["sys/policy"] = sysPolicyHandler
		}
	}

	return ConfigWalker{
		HandlerMap: handlerMap,
		Client: client,
		ConfigDir: path.Clean(docPath),
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
		if v == "*" {
			// not a real path, just used to store our generic handler
			continue
		}
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
	if f == nil {
		log.Printf("f nil for path %q", path)
		return nil
	}
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
	if ok {
		log.Printf("Processing %s using %T handler", path, handler)
		return handler.PutPoliciesFromDir(path)
	}

	// At this point, we have a directory, which has no handler assigned to itself or any parent
	log.Printf("Processing path %s using generic handler", relPath)
	genericHandler := cw.HandlerMap["*"]
	return genericHandler.PutPoliciesFromDir(path)
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

