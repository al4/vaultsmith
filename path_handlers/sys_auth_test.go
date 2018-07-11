package path_handlers

import (
	"log"
	vaultApi "github.com/hashicorp/vault/api"
	"testing"
	"strings"
	"os"
	"path/filepath"
	"github.com/starlingbank/vaultsmith/vault"
)

// calculate path to test fixtures (example/)
func examplePath() string {
	wd, _ := os.Getwd()
	pathArray := strings.Split(wd, string(os.PathSeparator))
	pathArray = pathArray[:len(pathArray) - 1]  // trims "internal"
	path := append(pathArray, "example")
	return strings.Join(path, string(os.PathSeparator))
}

func TestSysAuth_EnsureAuth(t *testing.T) {
	// Not terribly testable as it doesn't return anything we can assert against
	client := &vault.MockClient{}
	sh, err := NewSysAuthHandler(client, PathHandlerConfig{})
	if err != nil {
		log.Fatalf("Failed to create SysAuth: %s", err)
	}

	enableOpts := vaultApi.EnableAuthOptions{}
	err = sh.EnsureAuth("foo", enableOpts)
	if err != nil {
		log.Fatalf("Error calling EnsureAuth: %s", err)
	}
}

func TestSysAuth_PutPoliciesFromDir_Empty(t *testing.T) {
	// Should do nothing without error
	client := &vault.MockClient{}
	sh, err := NewSysAuthHandler(client, PathHandlerConfig{})
	if err != nil {
		log.Fatalf("Failed to create SysAuth: %s", err)
	}
	err = sh.PutPoliciesFromDir("")
	if err != nil {
		log.Fatalf("Expected nil, got error %s", err.Error())
	}
}

func TestSysAuth_PutPoliciesFromDir_Example(t *testing.T) {
	client := &vault.MockClient{}
	sh, err := NewSysAuthHandler(client, PathHandlerConfig{
		DocumentPath: examplePath(),
	})
	if err != nil {
		log.Fatalf("Failed to create SysAuth: %s", err)
	}

	sysPath := filepath.Join(examplePath(), "sys/auth")
	err = sh.PutPoliciesFromDir(sysPath)

	if err != nil {
		log.Fatalf("Expected no error, got %q", err)
	}
}

func TestSysAuth_WalkFile(t *testing.T) {
	//client := &vaultClient.MockClient{}
	//sh, err := NewSysAuthHandler(client, "")
	//if err != nil {
	//	log.Fatalf("Failed to create SysAuth: %s", err)
	//}

}
