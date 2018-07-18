package path_handlers

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/starlingbank/vaultsmith/document"
	"github.com/starlingbank/vaultsmith/vault"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

// Required information to write a document to vault
type VaultDocument struct {
	path string
	data map[string]interface{}
}

// The generic handler simply writes the files to the path they are stored in
type GenericHandler struct {
	BaseHandler
}

func NewGenericHandler(c vault.Vault, config PathHandlerConfig) (*GenericHandler, error) {
	return &GenericHandler{
		BaseHandler: BaseHandler{
			client: c,
			config: config,
			log: log.WithFields(log.Fields{
				"handler": "Generic",
			}),
		},
	}, nil
}

func (gh *GenericHandler) walkFile(path string, f os.FileInfo, err error) error {
	if f == nil {
		log.Debugf("No path %q present, skipping", path)
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

	// common variables for write path
	apiPath, err := apiPath(gh.config.DocumentPath, path)
	if err != nil {
		return fmt.Errorf("could not determine relative path of %s to %s: %s",
			path, gh.config.DocumentPath, err)
	}

	for _, td := range templatedDocs {
		// parse our document data as json
		var data map[string]interface{}
		err = json.Unmarshal([]byte(td.Content), &data)
		if err != nil {
			return fmt.Errorf("failed to parse json from file %q: %s", path, err)
		}

		writePath := templatePath(apiPath, td.Name)

		doc := VaultDocument{
			path: writePath,
			data: data,
		}
		err := gh.ensureDoc(doc)
		if err != nil {
			return err
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
func (gh *GenericHandler) ensureDoc(doc VaultDocument) error {
	if applied, err := gh.isDocApplied(doc); err != nil {
		return fmt.Errorf("could not determine if %q is applied: %s", doc.path, err)
	} else if applied {
		gh.log.Debugf("Document %q already applied", doc.path)
		return nil
	}

	gh.log.Infof("Applying %q", doc.path)
	_, err := gh.client.Write(doc.path, doc.data)
	return err
}

// true if the document is on the server and matches the one configured
func (gh *GenericHandler) isDocApplied(doc VaultDocument) (bool, error) {
	secret, err := gh.client.Read(doc.path)
	if err != nil {
		// TODO assume not applied, but should handle specific errors differently
		gh.log.Errorf("error on client.Read, assuming doc %q not present: %s, please raise a "+
			"bug as this should be handled cleanly!", doc.path, err)
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
		if _, ok := mapB[key]; !ok {
			return false // not present at all
		}
		if reflect.DeepEqual(mapA[key], mapB[key]) {
			continue // value the same, skip further checks for this key
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
		gh.log.Debugf("%q not equal; %+v(%T) != %+v(%T)", key, mapA[key], mapA[key], mapB[key], mapB[key])
		return false
	}
	return true
}

func (gh *GenericHandler) RemoveUndeclaredDocuments() (removed []string, err error) {
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
		log.Errorf("Unhandled type %T in firstElementEqual(), needs to be added to the "+
			"switch statement; please raise a bug", t)
		return false
	}

	return false
}

// Determine whether a string ttl is equal to an int ttl
func IsTtlEquivalent(ttlA interface{}, ttlB interface{}) bool {
	durA, err := convertToDuration(ttlA)
	if err != nil {
		log.Warnf("Could not parse %+v: %s", ttlA, err)
		return false
	}
	durB, err := convertToDuration(ttlB)
	if err != nil {
		log.Warnf("Could not convert %+v to duration: %s", ttlA, err)
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
