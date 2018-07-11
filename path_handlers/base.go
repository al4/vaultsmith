package path_handlers

import (
	vaultApi "github.com/hashicorp/vault/api"
	"os"
	"fmt"
	"bytes"
	"io"
	"log"
	"github.com/starlingbank/vaultsmith/vault"
)

type PathHandlerConfig struct {
	DocumentPath string  // path to the base of the vault documents
	Order int //
}

// A PathHandler takes a path and applies the policies within
type PathHandler interface {
	PutPoliciesFromDir(path string) error
	Order() int
}

type ValueMap map[string][]string

// Set of methods common to all PathHandlers
type BaseHandler struct {
	client            vault.Vault
	rootPath          string
	liveAuthMap       *map[string]*vaultApi.AuthMount
	configuredAuthMap *map[string]*vaultApi.AuthMount
	order             int  // order to process. Lower is earlier, with the exception of 0, which
							 // is processed after any others with a positive integer
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

func (h *BaseHandler) Order() int {
	return h.order
}