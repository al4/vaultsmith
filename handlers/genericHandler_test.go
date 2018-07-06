package handlers

import (
	"testing"
	"github.com/starlingbank/vaultsmith/vaultClient"
	"log"
	vaultApi "github.com/hashicorp/vault/api"
)

func TestGenericHandler_isDocApplied_true(t *testing.T) {
	testData := make(map[string]interface{})
	testData["testKey"] = "testValue"
	testDoc := Document{"test/path", testData}

	testSecret := vaultApi.Secret{
		Data: testData,
	}
	client := &vaultClient.MockVaultsmithClient{
		ReturnSecret: &testSecret,
	}

	gh, err := NewGenericHandler(client, "N/A", "N/A")
	if err != nil {
		log.Fatal("Failed to create generic handler")
	}

	gh.liveDocMap["test/path"] = testDoc
	gh.configuredDocMap["test/path"] = testDoc

	result, err := gh.isDocApplied(testDoc)
	if err != nil {
		log.Fatalf("Error calling isDocApplied: %s", err)
	}
	if ! result {
		log.Fatalf("Got false result, expected true")
	}
}
