package handlers

import (
	"log"
)

// A minimal handler which does nothing, mainly for testing
type DummyHandler struct {
	BasePathHandler
	client 				VaultsmithClient
	rootPath 			string
}

func NewDummyHandler(c VaultsmithClient, rootPath string) (*DummyHandler, error) {
	return &DummyHandler{
		client: c,
		rootPath: rootPath,
	}, nil
}

func (sh *DummyHandler) PutPoliciesFromDir(path string) error {
	log.Printf("Dummy handler got path: %s", path)
	return nil
}
