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
type vaultDocument struct {
	path       string
	data       map[string]interface{}
	sourceFile string
}

// The generic handler simply writes the files to the path they are stored in
type Generic struct {
	BaseHandler
	configuredDocMap map[string]vaultDocument
	removedDocMap    map[string]interface{}
}

func NewGeneric(client vault.Vault, config PathHandlerConfig) (*Generic, error) {
	return &Generic{
		BaseHandler: BaseHandler{
			client: client,
			config: config,
			log: log.WithFields(log.Fields{
				"handler": "Generic",
			}),
		},
		configuredDocMap: map[string]vaultDocument{},
		removedDocMap:    map[string]interface{}{},
	}, nil
}

func (gh *Generic) walkFile(path string, f os.FileInfo, err error) error {
	logger := gh.log.WithFields(log.Fields{
		"path":  path,
		"error": err,
		"file":  f,
	})
	if f == nil {
		logger.Debugf("Path not present, skipping", path)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error finding %s: %s", path, err)
	}
	// not doing anything with dirs
	if f.IsDir() {
		return nil
	}

	tp, err := document.GenerateTemplateParams(gh.config.TemplateFile, gh.config.TemplateOverrides)
	if err != nil {
		return fmt.Errorf("could not generate template parameters: %s", err)
	}

	content, err := document.Read(path)
	if err != nil {
		return fmt.Errorf("error reading %q: %s", path, err)
	}
	td := &document.Template{
		FileName: f.Name(),
		Content:  content,
		Params:   tp,
	}

	templatedDocs, err := td.Render()
	if err != nil {
		return fmt.Errorf("failed to render document %q: %s", path, err)
	}

	// figure out where to write to
	apiDir, err := apiDir(gh.config.DocumentPath, path)
	if err != nil {
		return err
	}

	for _, td := range templatedDocs {
		// parse our document data as json
		var data map[string]interface{}
		err = json.Unmarshal([]byte(td.Content), &data)
		if err != nil {
			log.Debugf("Content:\n%s", data)
			return fmt.Errorf("failed to parse json from file %q: %s", path, err)
		}

		doc := vaultDocument{
			path:       filepath.Join(apiDir, td.Name),
			data:       data,
			sourceFile: f.Name(),
		}
		err := gh.ensureDoc(doc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gh *Generic) PutPoliciesFromDir(path string) error {
	// path must be a real file system path here, not the relative path to the document root
	err := filepath.Walk(path, gh.walkFile)
	if err != nil {
		return err
	}

	return gh.removeUndeclaredDocuments(path)
}

// Ensure the document is present and consistent
func (gh *Generic) ensureDoc(doc vaultDocument) error {
	logger := gh.log.WithFields(log.Fields{
		"path":       doc.path,
		"sourceFile": doc.sourceFile,
	})
	gh.configuredDocMap[doc.path] = doc

	if applied, err := gh.isDocApplied(doc); err != nil {
		return fmt.Errorf("could not determine if %q is applied: %s", doc.path, err)
	} else if applied {
		logger.Debugf("Document already applied")
		return nil
	}

	logger.Infof("Applying document")
	_, err := gh.client.Write(doc.path, doc.data)
	return err
}

// true if the document is on the server and matches the one configured
func (gh *Generic) isDocApplied(doc vaultDocument) (bool, error) {
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
func (gh *Generic) areKeysApplied(mapA map[string]interface{}, mapB map[string]interface{}) bool {
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
			if isTtlEquivalent(mapA[key], mapB[key]) {
				continue
			}
		}
		// covers cases such as "policy" == ["policy]
		// logic is a bit scary, see function documentation
		if isSliceEquivalent(mapA[key], mapB[key]) {
			continue
		}
		//gh.log.Debugf("%q not equal; %+v(%T) != %+v(%T)", key, mapA[key], mapA[key], mapB[key], mapB[key])
		return false
	}
	return true
}

// Remove documents that are not declared
// Note; only the configured path for this handler is affected
func (gh *Generic) removeUndeclaredDocuments(path string) (err error) {
	err = filepath.Walk(path, gh.removalWalk)
	return
}

func (gh *Generic) removalWalk(path string, f os.FileInfo, err error) error {
	if !f.IsDir() {
		return nil
	}
	apiPath, err := apiPath(gh.config.DocumentPath, path)
	if err != nil {
		return err
	}

	secret, err := gh.client.List(apiPath)
	if err != nil {
		return err
	}
	if secret == nil {
		// path missing or nothing in it; either way, skip
		return err
	}
	var keys []interface{}
	if v, ok := secret.Data["keys"]; !ok {
		// List() returns a vault Secret object with a Data map, and sub-directories in a "keys"
		// field of that map
		return fmt.Errorf("secret data did not contain field 'keys'")
	} else if k, ok := v.([]interface{}); ok {
		// cast to array, as vault secret data can be arbitrary types
		keys = k
	} else {
		return fmt.Errorf("could not cask keys value '%+v' as an array", v)
	}

	for k := range keys {
		docPath := strings.Join([]string{apiPath, keys[k].(string)}, "/")
		if _, ok := gh.configuredDocMap[docPath]; ok {
			// configured, leave it alone
			continue
		}
		logger := gh.log.WithFields(log.Fields{"docPath": docPath})

		logger.Info("Removing document")
		_, err := gh.client.Delete(docPath)
		if err != nil {
			return err
		}
		gh.removedDocMap[docPath] = true
	}

	return nil
}

func (gh *Generic) Order() int {
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
func isTtlEquivalent(ttlA interface{}, ttlB interface{}) bool {
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
