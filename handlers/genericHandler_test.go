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

	returnSecret := vaultApi.Secret{
		Data: testData,
	}
	client := &vaultClient.MockVaultsmithClient{
		ReturnSecret: &returnSecret,
	}

	gh, err := NewGenericHandler(client, "N/A", "N/A")
	if err != nil {
		log.Fatal("Failed to create generic handler")
	}

	result, err := gh.isDocApplied(testDoc)
	if err != nil {
		log.Fatalf("Error calling isDocApplied: %s", err)
	}
	if ! result {
		log.Fatalf("Got false result, expected true")
	}
}

func TestGenericHandler_isDocApplied_falseValue(t *testing.T) {
	testDataA := make(map[string]interface{})
	testDataB := make(map[string]interface{})

	testDataA["testKey"] = "testValue"
	testDoc := Document{"test/path", testDataA}

	testDataB["testKey"] = "otherValue"
	returnSecret := vaultApi.Secret{
		Data: testDataB,
	}
	client := &vaultClient.MockVaultsmithClient{
		ReturnSecret: &returnSecret,
	}

	gh, err := NewGenericHandler(client, "N/A", "N/A")
	if err != nil {
		log.Fatal("Failed to create generic handler")
	}

	result, err := gh.isDocApplied(testDoc)
	if err != nil {
		log.Fatalf("Error calling isDocApplied: %s", err)
	}
	if result {
		log.Fatalf("Got true result, expected false")
	}
}

func TestGenericHandler_areKeysApplied_true(t *testing.T) {
	client := &vaultClient.MockVaultsmithClient{}

	gh, err := NewGenericHandler(client, "N/A", "N/A")
	if err != nil {
		log.Fatal("Failed to create generic handler")
	}

	testDataA := make(map[string]interface{})
	testDataB := make(map[string]interface{})

	testDataA["testKey"] = "testValue"

	testDataB["testKey"] = "testValue"
	testDataB["otherKey"] = "otherValue"  // extra values are OK, we only care if the defined ones are present

	r := gh.areKeysApplied(testDataA, testDataB)
	if ! r {
		log.Fatal("Expected areKeysApplied to return true")
	}

}

func TestGenericHandler_areKeysApplied_false(t *testing.T) {
	client := &vaultClient.MockVaultsmithClient{}

	gh, err := NewGenericHandler(client, "N/A", "N/A")
	if err != nil {
		log.Fatal("Failed to create generic handler")
	}

	testDataA := make(map[string]interface{})
	testDataB := make(map[string]interface{})

	testDataA["testKey"] = "testValue"
	testDataA["otherKey"] = "otherValue" // this is not OK because it will not be present in B

	testDataB["testKey"] = "testValue"

	r := gh.areKeysApplied(testDataA, testDataB)
	if r {
		log.Fatal("Expected areKeysApplied to return false")
	}

}
