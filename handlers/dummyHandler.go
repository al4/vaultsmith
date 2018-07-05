package handlers

import (
	"log"
	"github.com/starlingbank/vaultsmith/vaultClient"
)

// A minimal handler which does nothing, mainly for testing
type DummyHandler struct {
	BasePathHandler
	client 				vaultClient.VaultsmithClient
	rootPath 			string
}

func NewDummyHandler(c vaultClient.VaultsmithClient, rootPath string) (*DummyHandler, error) {
	return &DummyHandler{
		client: c,
		rootPath: rootPath,
	}, nil
}

func (sh *DummyHandler) PutPoliciesFromDir(path string) error {
	log.Printf("Dummy handler got path: %s", path)
	return nil
}
