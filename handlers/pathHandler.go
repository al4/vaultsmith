package handlers

import (
	vaultApi "github.com/hashicorp/vault/api"
	"os"
	"fmt"
	"bytes"
	"io"
	"log"
	"github.com/starlingbank/vaultsmith/vaultClient"
)

// A PathHandler takes a path and applies the policies within
type PathHandler interface {
	PutPoliciesFromDir(path string) error
	Order() int
}

type ValueMap map[string][]string

// Base set of methods common to all PathHandlers
type BasePathHandler struct {
	client 				vaultClient.VaultsmithClient
	rootPath 			string
	liveAuthMap 		*map[string]*vaultApi.AuthMount
	configuredAuthMap 	*map[string]*vaultApi.AuthMount
	order				int  // order to process. 0 is treated as higher than (after) any positive int
}

func (h *BasePathHandler) readFile(path string) (string, error) {
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

func (h *BasePathHandler) Order() int {
	return h.order
}