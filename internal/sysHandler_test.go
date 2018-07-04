package internal

import (
	"log"
	"github.com/starlingbank/vaultsmith/mocks"
	vaultApi "github.com/hashicorp/vault/api"
	"testing"
	"strings"
)


func TestEnsureAuth(t *testing.T) {
	// Not terribly testable as it doesn't return anything we can assert against
	client := &mocks.MockVaultsmithClient{}
	sh, err := NewSysHandler(client, "")
	if err != nil {
		log.Fatalf("Failed to create SysHandler: %s", err)
	}

	enableOpts := vaultApi.EnableAuthOptions{}
	err = sh.EnsureAuth("foo", enableOpts)
	if err != nil {
		log.Fatalf("Error calling EnsureAuth: %s", err)
	}
}

func TestPutPoliciesFromEmptyDir(t *testing.T) {
	client := &mocks.MockVaultsmithClient{}
	sh, err := NewSysHandler(client, "")
	if err != nil {
		log.Fatalf("Failed to create SysHandler: %s", err)
	}
	err = sh.PutPoliciesFromDir("")
	if err == nil {
		log.Fatal("Expected error, got nil")
	}
	if ! strings.Contains(err.Error(), "FileInfo") {
		log.Fatalf("Expected error about nil FileInfo, got '%s'", err.Error())
	}

}

func TestPutPoliciesFromExampleDir(t *testing.T) {
	client := &mocks.MockVaultsmithClient{}
	sh, err := NewSysHandler(client, "./example")
	if err != nil {
		log.Fatalf("Failed to create SysHandler: %s", err)
	}
	err = sh.PutPoliciesFromDir("example/auth")
	log.Println(err)

}

func TestSysHandlerWalkFile(t *testing.T) {
	//client := &mocks.MockVaultsmithClient{}
	//sh, err := NewSysHandler(client, "")
	//if err != nil {
	//	log.Fatalf("Failed to create SysHandler: %s", err)
	//}

}
