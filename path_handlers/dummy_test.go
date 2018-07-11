package path_handlers

import (
	"testing"
	"log"
	"github.com/starlingbank/vaultsmith/vaultClient"
)

func TestNewDummyHandler(t *testing.T) {
	c := &vaultClient.MockVaultsmithClient{}
	_, err := NewDummyHandler(c, "example", 0)
	if err != nil {
		log.Fatalf("Failed to create dummy handler: %s", err.Error())
	}
}