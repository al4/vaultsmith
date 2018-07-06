package handlers

import (
	"testing"
	"github.com/starlingbank/vaultsmith/vaultClient"
	"log"
)

func TestGenericHandler_isDocApplied_true(t *testing.T) {
	client := &vaultClient.MockVaultsmithClient{}
	gh, err := NewGenericHandler(client, "", "")
	if err != nil {
		log.Fatal("Failed to create generic handler")
	}

	testData := make(map[string]interface{})
	testData["testKey"] = "testValue"
	testDoc := Document{"test/path", testData}

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
