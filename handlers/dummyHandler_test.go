package handlers

import (
	"testing"
	"github.com/starlingbank/vaultsmith/mocks"
	"log"
)

func TestNewDummyHandler(t *testing.T) {
	c := &mocks.MockVaultsmithClient{}
	_, err := NewDummyHandler(c, "example")
	if err != nil {
		log.Fatalf("Failed to create dummy hanlder: %s", err.Error())
	}
}