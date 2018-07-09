package main

import (
	"log"
	"strings"
	"testing"
	"fmt"
	"github.com/starlingbank/vaultsmith/vaultClient"
	"github.com/starlingbank/vaultsmith/config"
)

func TestRunWhenVaultNotListening(t *testing.T) {
	conf := &config.VaultsmithConfig{
		VaultRole: "ValidRole",
	}
	mockClient := new(vaultClient.MockVaultsmithClient)
	conf.VaultRole = "ConnectionRefused"
	mockClient.On("Authenticate", conf.VaultRole)

	err := Run(mockClient, conf)
	if err == nil {
		log.Fatal("Expected error, got nil")
	}
	if ! strings.Contains(err.Error(), "failed authenticating with Vault:") {
		log.Fatal("bad failure message")
	}
	if ! strings.Contains(err.Error(), "connection refused") {
		log.Fatal("bad reason message")
	}
}

func TestRunWhenRoleIsInvalid(t *testing.T) {
	conf := &config.VaultsmithConfig{
		VaultRole: "ValidRole",
	}
	mockClient := new(vaultClient.MockVaultsmithClient)
	conf.VaultRole = "InvalidRole"
	mockClient.On("Authenticate", conf.VaultRole)

	err := Run(mockClient, conf)
	if err == nil {
		log.Fatal("Expected error, got nil")
	}
	if ! strings.Contains(err.Error(), "failed authenticating with Vault:") {
		log.Fatalf("bad failure message '%s'", err.Error())
	}

	if ! strings.Contains(err.Error(), fmt.Sprintf("entry for role %s not found", conf.VaultRole)) {
		log.Fatalf("bad reason message '%s'", err.Error())
	}
}
