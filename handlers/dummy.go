package handlers

import (
	"log"
	"github.com/starlingbank/vaultsmith/vaultClient"
)

// A minimal handler which does nothing, mainly for testing
type Dummy struct {
	BasePathHandler
	client 				vaultClient.VaultsmithClient
	rootPath 			string  // path to handle
	order				int
}

func NewDummyHandler(c vaultClient.VaultsmithClient, rootPath string, order int) (*Dummy, error) {
	return &Dummy{
		client: c,
		rootPath: rootPath,
		order: order,
	}, nil
}

func (h *Dummy) PutPoliciesFromDir(path string) error {
	log.Printf("Dummy handler got path: %s", path)
	return nil
}

func (h *Dummy) Order() int {
	return h.order
}
