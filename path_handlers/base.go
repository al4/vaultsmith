package path_handlers

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/starlingbank/vaultsmith/vault"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type PathHandlerConfig struct {
	DocumentPath string // path to the base of the vault documents
	Order        int    // order to process (lower int is earlier, except 0 is last)
	MappingFile  string
}

// A PathHandler takes a path and applies the policies within
type PathHandler interface {
	PutPoliciesFromDir(path string) error
	Order() int
	Name() string
}

type ValueMap map[string][]string

// Set of methods common to all PathHandlers
type BaseHandler struct {
	client   vault.Vault
	config   PathHandlerConfig
	rootPath string
	order    int // order to process. Lower is earlier, with the exception of 0, which is
	// processed after any others with a positive integer
	name string
	log  *log.Entry
}

func (h *BaseHandler) Name() string {
	return h.name
}

func (h *BaseHandler) Order() int {
	return h.order
}

func (h *BaseHandler) readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("error opening file: %s", err)
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)

	if err != nil {
		log.Fatal(fmt.Sprintf("error reading from buffer: %s", err))
	}

	data := buf.String()

	return data, nil
}

// Return the vault api path for this rendered template, given the filesystem path
// Basically, relative path to the root, sans extensions
func apiPath(rootPath string, filePath string) (apiPath string, err error) {
	relPath, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		return apiPath, fmt.Errorf("could not determine relative path of %s to %s: %s",
			filePath, rootPath, err)
	}

	dir, file := filepath.Split(relPath)
	// strip any extensions
	fileName := strings.Split(file, ".")[0]

	// TODO This assumes OS path separator is '/'...
	return filepath.Join(dir, fileName), err
}

// Return the template path
// This is the ApiPath with the template name
func templatePath(apiPath string, name string) (templatePath string) {
	if name != "" {
		return fmt.Sprintf("%s_%s", apiPath, name)
	} else {
		return apiPath
	}
}
