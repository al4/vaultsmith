package path_handlers

import (
	log "github.com/sirupsen/logrus"
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
			log: log.WithFields(log.Fields{
				"handler": "generic",
			}),
		},
	}, nil
}

func (h *Dummy) PutPoliciesFromDir(path string) error {
	h.log.Debugf("Dummy handler got path: %s", path)
	return nil
}

func (h *Dummy) Order() int {
	return h.order
}
