package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/starlingbank/vaultsmith/config"
	"github.com/starlingbank/vaultsmith/vault"
	"strings"
	"testing"
)

func TestRunWhenVaultNotListening(t *testing.T) {
	conf := config.VaultsmithConfig{
		VaultRole: "ValidRole",
	}
	mockClient := new(vault.MockClient)
	conf.VaultRole = "ConnectionRefused"
	mockClient.On("Authenticate", conf.VaultRole)

	err := Run(mockClient, conf)
	if err == nil {
		log.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed authenticating with Vault:") {
		log.Fatal("bad failure message")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		log.Fatal("bad reason message")
	}
}

func TestRunWhenRoleIsInvalid(t *testing.T) {
	conf := config.VaultsmithConfig{
		VaultRole: "ValidRole",
	}
	mockClient := new(vault.MockClient)
	conf.VaultRole = "InvalidRole"
	mockClient.On("Authenticate", conf.VaultRole)

	err := Run(mockClient, conf)
	if err == nil {
		log.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed authenticating with Vault:") {
		t.Errorf("bad failure message '%s'", err.Error())
	}

	if !strings.Contains(err.Error(), fmt.Sprintf("entry for role %s not found", conf.VaultRole)) {
		t.Errorf("bad reason message '%s'", err.Error())
	}
}
