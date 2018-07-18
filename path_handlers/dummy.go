package path_handlers

import (
	"github.com/starlingbank/vaultsmith/vault"
)

// A minimal handler which does nothing, mainly for testing
type Dummy struct {
	BaseHandler
	client   vault.Vault
	rootPath string // path to handle
}

func NewDummyHandler(c vault.Vault, rootPath string, order int) (*Dummy, error) {
	return &Dummy{
		BaseHandler: BaseHandler{
			client:   c,
			rootPath: rootPath,
			order:    order,
			name:     "Dummy",
		},
	}, nil
}

func (h *Dummy) PutPoliciesFromDir(path string) error {
	//log.Printf("Dummy handler got path: %s", path)
	return nil
}

func (h *Dummy) Order() int {
	return h.order
}
