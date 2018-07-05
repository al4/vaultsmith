package handlers

import (
	"testing"
	"log"
	"github.com/starlingbank/vaultsmith/vaultClient"
)

func TestNewDummyHandler(t *testing.T) {
	c := &vaultClient.MockVaultsmithClient{}
	_, err := NewDummyHandler(c, "example")
	if err != nil {
		log.Fatalf("Failed to create dummy hanlder: %s", err.Error())
	}
}