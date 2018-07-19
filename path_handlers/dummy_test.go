package path_handlers

import (
	"github.com/starlingbank/vaultsmith/vault"
	"testing"
)

func TestNewDummyHandler(t *testing.T) {
	c := &vault.MockClient{}
	_, err := NewDummyHandler(c, "example", 0)
	if err != nil {
		t.Errorf("Failed to create dummy handler: %s", err.Error())
	}
}
