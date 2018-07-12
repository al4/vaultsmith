package path_handlers

import (
	"github.com/starlingbank/vaultsmith/vault"
	"log"
	"os"
	"fmt"
	"path/filepath"
	"strings"
	"encoding/json"
	"reflect"
	"github.com/starlingbank/vaultsmith/document"
	"time"
)

// Required information to write a document to vault
type VaultDocument struct {
	name string
	path string
	data map[string]interface{}
}

// The generic handler simply writes the files to the path they are stored in
type GenericHandler struct {
	BaseHandler
	client      vault.Vault
	config		PathHandlerConfig
}

func NewGenericHandler(c vault.Vault, config PathHandlerConfig) (*GenericHandler, error) {
	return &GenericHandler{
		client: c,
		config: config,
	}, nil
}

func (gh *GenericHandler) walkFile(path string, f os.FileInfo, err error) error {
	if f == nil {
		log.Printf("No path %q present, skipping", path)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading %q: %s", path, err)
	}
	// not doing anything with dirs
	if f.IsDir() {
		return nil
	}

	// getting file contents
	td, err := document.NewTemplate(path, gh.config.MappingFile)
	if err != nil {
		return fmt.Errorf("failed to instantiate TemplateDocument: %s", err)
	}

	templatedDocs, err := td.Render()
	if err != nil {
		return fmt.Errorf("failed to render document %q: %s", path, err)
	}

	// variables for write path
	relPath, err := filepath.Rel(gh.config.DocumentPath, path)
	if err != nil {
		return fmt.Errorf("could not determine relative path of %s to %s: %s",
			path, gh.config.DocumentPath, err)
	}
	dir, file := filepath.Split(relPath)
	fileName := strings.Split(file , ".")[0]

	for name, content := range templatedDocs {
		// parse our document data as json
		var data map[string]interface{}
		err = json.Unmarshal([]byte(content), &data)
		if err != nil {
			return fmt.Errorf("failed to parse json from file %q: %s", path, err)
		}

		var docName string
		if name == "" { // this is the only instance, or purposely unlabelled
			docName = fileName
		} else { // need to write each one to a separate path
			docName = fmt.Sprintf("%s_%s", fileName, name)
		}
		writePath := filepath.Join(dir, docName)
		log.Printf("Applying %q", docName)

		doc := VaultDocument{
			name: docName,
			path: writePath,
			data: data,
		}
		return gh.EnsureDoc(doc)
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
func (gh *GenericHandler) EnsureDoc(doc VaultDocument) error {
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
func (gh *GenericHandler) isDocApplied(doc VaultDocument) (bool, error) {
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
		if isSliceEquivalent(mapA[key], mapB[key]) {
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

// Determine whether an array is logically equivalent as far as Vault is concerned.
// e.g. [policy] == policy
func isSliceEquivalent(a interface{}, b interface{}) (equivalent bool) {
	if reflect.TypeOf(a).Kind() == reflect.TypeOf(b).Kind() {
		// just compare directly if type is the same
		return reflect.DeepEqual(a, b)
	}

	if reflect.TypeOf(a).Kind() == reflect.Slice {
		// b must not be a slice, compare a[0] to it
		return firstElementEqual(a, b)
	}

	if reflect.TypeOf(b).Kind() == reflect.Slice {
		// a must not be a slice, compare b[0] to it
		return firstElementEqual(b, a)
	}

	return false
}

// Return true if value is equal to the first item in slice
func firstElementEqual(slice interface{}, value interface{}) bool {
	switch t := slice.(type) {
	case []string:
		if t[0] == value && len(t) == 1 {
			return true
		}
	case []int:
		if t[0] == value && len(t) == 1 {
			return true
		}
	case []interface{}:
		s := reflect.ValueOf(t)
		var val interface{}
		for i := 0; i < s.Len(); i++ {
			if i > 0 { // length > 1, cannot be equivalent
				return false
			}
			if i == 0 {
				// This is a little scary in a strongly typed context, as we're parsing everything
				// as a string. But in the context of vault API responses it should be OK...
				val = fmt.Sprintf("%v", s.Index(i))
			}
		}
		if val == value {
			return true
		}
	default:
		log.Fatalf("Unhandled type %T, please add this to the switch statement", t)
	}

	return false
}

// Determine whether a string ttl is equal to an int ttl
func IsTtlEquivalent(ttlA interface{}, ttlB interface{}) bool {
	durA, err := convertToDuration(ttlA)
	if err != nil {
		log.Printf("WARN: Error parsing %+v: %s", ttlA, err)
		return false
	}
	durB, err := convertToDuration(ttlB)
	if err != nil {
		log.Printf("WARN: Error converting %+v to duration: %s", ttlA, err)
		return false
	}

	if durA == durB {
		return true
	}

	return false
}

// convert x to time.Duration. if x is an integer, we assume it is in seconds
func convertToDuration(x interface{}) (time.Duration, error) {
	var duration time.Duration
	var err error

	switch x.(type) {
	case string:
		duration, err = time.ParseDuration(x.(string))
		if err != nil {
			return 0, fmt.Errorf("%q can't be parsed as duration", x)
		}
	case int64:
		duration = time.Duration(x.(int64)) * time.Second
	case int:
		duration = time.Duration(int64(x.(int))) * time.Second
	case json.Number:
		i, err := x.(json.Number).Int64()
		if err != nil {
			return 0, fmt.Errorf("could not parse %+v as json number: %s", x, err.Error())
		}
		duration = time.Duration(i) * time.Second
	default:
		return 0, fmt.Errorf("type of '%+v' not handled", reflect.TypeOf(x))
	}

	return duration, nil
}

