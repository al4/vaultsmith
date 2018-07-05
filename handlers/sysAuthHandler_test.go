package handlers

import (
	"log"
	"github.com/starlingbank/vaultsmith/mocks"
	vaultApi "github.com/hashicorp/vault/api"
	"testing"
	"strings"
	"os"
	"path/filepath"
)

// calculate path to test fixtures (example/)
func examplePath() string {
	wd, _ := os.Getwd()
	pathArray := strings.Split(wd, string(os.PathSeparator))
	pathArray = pathArray[:len(pathArray) - 1]  // trims "internal"
	path := append(pathArray, "example")
	return strings.Join(path, string(os.PathSeparator))
}

func TestEnsureAuth(t *testing.T) {
	// Not terribly testable as it doesn't return anything we can assert against
	client := &mocks.MockVaultsmithClient{}
	sh, err := NewSysAuthHandler(client, "")
	if err != nil {
		log.Fatalf("Failed to create SysAuthHandler: %s", err)
	}

	enableOpts := vaultApi.EnableAuthOptions{}
	err = sh.EnsureAuth("foo", enableOpts)
	if err != nil {
		log.Fatalf("Error calling EnsureAuth: %s", err)
	}
}

func TestPutPoliciesFromEmptyDir(t *testing.T) {
	client := &mocks.MockVaultsmithClient{}
	sh, err := NewSysAuthHandler(client, "")
	if err != nil {
		log.Fatalf("Failed to create SysAuthHandler: %s", err)
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
	sh, err := NewSysAuthHandler(client, examplePath())
	if err != nil {
		log.Fatalf("Failed to create SysAuthHandler: %s", err)
	}

	sysPath := filepath.Join(examplePath(), "sys")
	err = sh.PutPoliciesFromDir(sysPath)

	if err != nil {
		log.Fatalf("Expected no error, got: %s", err)
	}
}

func TestSysHandlerWalkFile(t *testing.T) {
	//client := &mocks.MockVaultsmithClient{}
	//sh, err := NewSysAuthHandler(client, "")
	//if err != nil {
	//	log.Fatalf("Failed to create SysAuthHandler: %s", err)
	//}

}
