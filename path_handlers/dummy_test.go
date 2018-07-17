package path_handlers

import (
	log "github.com/sirupsen/logrus"
	"github.com/starlingbank/vaultsmith/vault"
	"testing"
)

func TestNewDummyHandler(t *testing.T) {
	c := &vault.MockClient{}
	_, err := NewDummyHandler(c, "example", 0)
	if err != nil {
		log.Fatalf("Failed to create dummy handler: %s", err.Error())
	}
}
