package internal

import (
	"log"
	"github.com/starlingbank/vaultsmith/mocks"
	vaultApi "github.com/hashicorp/vault/api"
	"testing"
)


func TestEnsureAuth(t *testing.T) {
	// Not terribly testable as it doesn't return anything we can assert against
	client := &mocks.MockVaultsmithClient{}
	sh, err := NewSysHandler(client, "")
	if err != nil {
		log.Fatalf("failed to create SysHandler (using mock client): %s", err)
	}

	enableOpts := vaultApi.EnableAuthOptions{ }
	err = sh.EnsureAuth("foo", enableOpts)
	if err != nil {
		log.Fatalf("Error calling EnsureAuth: %s", err)
	}
}

