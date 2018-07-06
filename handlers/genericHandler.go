package handlers

import (
	"github.com/starlingbank/vaultsmith/vaultClient"
	"log"
	"os"
	"fmt"
	"path/filepath"
	"strings"
	"encoding/json"
	"reflect"
)

type Document struct {
	path string
	data map[string]interface{}
}

// The generic handler simply writes the files to the path they are stored in
type GenericHandler struct {
	BasePathHandler
	client 				vaultClient.VaultsmithClient
	rootPath 			string  // Where we walk from.
	globalRootPath		string  // The top level config directory. We need this as the relative path
								// is used to determine the vault write path.
}

func NewGenericHandler(c vaultClient.VaultsmithClient, globalRootPath string, rootPath string) (*GenericHandler, error) {
	return &GenericHandler{
		client: c,
		globalRootPath: globalRootPath,
		rootPath: rootPath,
	}, nil
}

func (gh *GenericHandler) walkFile(path string, f os.FileInfo, err error) error {
	if f == nil {
		return fmt.Errorf("got nil FileInfo for %q, error: '%s'", path, err.Error())
	}
	if err != nil {
		return fmt.Errorf("error reading %q: %s", path, err)
	}
	// not doing anything with dirs
	if f.IsDir() {
		return nil
	}

	log.Printf("Applying %s\n", path)

	// getting file contents
	fileContents, err := gh.readFile(path)
	if err != nil {
		return err
	}
	var data map[string]interface{}
	err = json.Unmarshal([]byte(fileContents), &data)
	if err != nil {
		return fmt.Errorf("failed to parse json from file %q: %s", path, err)
	}

	// determine write path
	relPath, err := filepath.Rel(gh.globalRootPath, path)
	if err != nil {
		return fmt.Errorf("could not determine relative path of %s to %s: %s", path, gh.globalRootPath, err)
	}
	dir, file := filepath.Split(relPath)
	fileName := strings.Split(file, ".")[0]  // sans extension
	writePath := filepath.Join(dir, fileName)
	log.Printf("writePath: %s", writePath)

	doc := Document{
		path: writePath,
		data: data,
	}
	err = gh.EnsureDoc(doc)
	if err != nil {
		return fmt.Errorf("failed to ensure document %q is applied: %s", path, err)
	}

	return nil
}

func (gh *GenericHandler) PutPoliciesFromDir(path string) error {
	err := filepath.Walk(path, gh.walkFile)
	if err != nil {
		return err
	}

	_, err = gh.RemoveUndeclaredDocuments()
	return err
}

// Ensure the document is present and consistent
func (gh *GenericHandler) EnsureDoc(doc Document) error {
	applied, err := gh.isDocApplied(doc)
	if err != nil {
		return fmt.Errorf("could not determine if %q is applied: %s", doc.path, err)
	}

	if applied {
		log.Printf("Document %s already applied", doc.path)
		return nil
	}

	log.Printf("Writing %q to server", doc.path)
	_, err = gh.client.Write(doc.path, doc.data)
	return err
}

// true if the document is on the server and matches the one configured
func (gh *GenericHandler) isDocApplied(doc Document) (bool, error) {
	secret, err := gh.client.Read(doc.path)
	if err != nil {
		// assume not applied, but what error?
		// TODO only mask "missing" errors
		log.Printf("TODO: error is %s", err)
		return false, nil
	}

	if secret == nil || secret.Data == nil {
		return false, nil
	}
	log.Printf("%+v", secret.Data)

	if reflect.DeepEqual(doc.data, secret.Data) {
		return true, nil
	}
	return false, nil
}

func (gh *GenericHandler) RemoveUndeclaredDocuments() (removed []string, err error){
	// TODO implement me
	return
}

