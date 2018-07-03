package internal

import (
	"testing"
	"log"
	"github.com/starlingbank/vaultsmith/mocks"
)

func TestVaultClient(t *testing.T) {
	c := mocks.MockVaultsmithClient{}
	sh, err := NewSysHandler(&c, "")
	if err != nil {
		log.Fatalf("could not create dummy SysHandler: %s", err)
	}
	log.Println(sh)
}
