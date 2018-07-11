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
	"github.com/starlingbank/vaultsmith/templateDocument"
	"github.com/starlingbank/vaultsmith/config"
)

type GenericDocument struct {
	path string
	data map[string]interface{}
}

// The generic handler simply writes the files to the path they are stored in
type GenericHandler struct {
	BaseHandler
	client 				vaultClient.VaultsmithClient
	rootPath 			string  // Where we walk from
	globalRootPath		string  // The top level config directory. We need this as the relative path
								// is used to determine the vault write path.
	mappingFile string
}

func NewGenericHandler(c vaultClient.VaultsmithClient, config config.VaultsmithConfig, rootPath string) (*GenericHandler, error) {
	return &GenericHandler{
		client: c,
		globalRootPath: config.ConfigDir,
		rootPath: rootPath,
		mappingFile: config.TemplateFile,
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
	td, err := templateDocument.NewTemplatedDocument(path, gh.mappingFile)
	if err != nil {
		return fmt.Errorf("failed to instantiate TemplateDocument: %s", err)
	}

	templatedDocs, err := td.Render()
	if err != nil {
		return fmt.Errorf("failed to render document %q: %s", path, err)
	}

	for _, content := range templatedDocs {
		// this function now feels overloaded
		var data map[string]interface{}
		err = json.Unmarshal([]byte(content), &data)
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

		doc := GenericDocument{
			path: writePath,
			data: data,
		}
		err = gh.EnsureDoc(doc)
		if err != nil {
			return fmt.Errorf("failed to ensure document %q is applied: %s", path, err)
		}
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
func (gh *GenericHandler) EnsureDoc(doc GenericDocument) error {
	applied, err := gh.isDocApplied(doc)
	if err != nil {
		return fmt.Errorf("could not determine if %q is applied: %s", doc.path, err)
	}

	if applied {
		log.Printf("Document %q already applied", doc.path)
		return nil
	}

	log.Printf("Writing %q to server", doc.path)
	_, err = gh.client.Write(doc.path, doc.data)
	return err
}

// true if the document is on the server and matches the one configured
func (gh *GenericHandler) isDocApplied(doc GenericDocument) (bool, error) {
	secret, err := gh.client.Read(doc.path)
	if err != nil {
		// TODO assume not applied, but should handle specific errors differently
		log.Printf("TODO: error on client.Read, assuming doc not present (%s)", err)
		return false, nil
	}

	if secret == nil || secret.Data == nil {
		return false, nil
	}

	return gh.areKeysApplied(doc.data, secret.Data), nil
}

// Ensure all key/value pairs in mapA are present and consistent in mapB
// extra keys in remoteMap are ignored
func (gh *GenericHandler) areKeysApplied(mapA map[string]interface{}, mapB map[string]interface{}) bool {
	for key := range mapA {
		if _, ok := mapB[key]; ! ok {
			return false  // not present at all
		}
		if reflect.DeepEqual(mapA[key], mapB[key]) {
			continue  // value the same, skip further checks for this key
		}

		// this is a bit more complicated, thanks to ttls and bundling into arrays :(
		if strings.Contains(key, "ttl") {
			// check if the ttls are equivalent
			if IsTtlEquivalent(mapA[key], mapB[key]) {
				continue
			}
		}
		// covers cases such as "policy" == ["policy]
		// logic is a bit scary, see function documentation
		if IsSliceEquivalent(mapA[key], mapB[key]) {
			continue
		}
		//log.Printf(" ## %q not equal; %+v(%T) != %+v(%T)", key, mapA[key], mapA[key], mapB[key], mapB[key])
		return false
	}
	return true
}

func (gh *GenericHandler) RemoveUndeclaredDocuments() (removed []string, err error){
	// TODO implement me
	return
}

func (gh *GenericHandler) Order() int {
	return gh.order
}
