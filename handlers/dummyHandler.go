package handlers

import (
	"log"
	"github.com/starlingbank/vaultsmith/vaultClient"
)

// A minimal handler which does nothing, mainly for testing
type DummyHandler struct {
	BasePathHandler
	client 				vaultClient.VaultsmithClient
	rootPath 			string  // path to handle
	order				int
}

func NewDummyHandler(c vaultClient.VaultsmithClient, rootPath string, order int) (*DummyHandler, error) {
	return &DummyHandler{
		client: c,
		rootPath: rootPath,
		order: order,
	}, nil
}

func (h *DummyHandler) PutPoliciesFromDir(path string) error {
	log.Printf("Dummy handler got path: %s", path)
	return nil
}

func (h *DummyHandler) Order() int {
	return h.order
}
