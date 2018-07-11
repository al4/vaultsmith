package path_handlers

import (
	"testing"
	"github.com/starlingbank/vaultsmith/vault"
	"log"
	vaultApi "github.com/hashicorp/vault/api"
	"encoding/json"
)

func TestGeneric_isDocApplied_true(t *testing.T) {
	testData := make(map[string]interface{})
	testData["testKey"] = "testValue"
	testDoc := GenericDocument{"test/path", testData}

	returnSecret := vaultApi.Secret{
		Data: testData,
	}
	client := &vault.MockClient{
		ReturnSecret: &returnSecret,
	}

	gh, err := NewGenericHandler(client, PathHandlerConfig{})
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

func TestGeneric_isDocApplied_falseValue(t *testing.T) {
	testDataA := make(map[string]interface{})
	testDataB := make(map[string]interface{})

	testDataA["testKey"] = "testValue"
	testDoc := GenericDocument{"test/path", testDataA}

	testDataB["testKey"] = "otherValue"
	returnSecret := vaultApi.Secret{
		Data: testDataB,
	}
	client := &vault.MockClient{
		ReturnSecret: &returnSecret,
	}

	gh, err := NewGenericHandler(client, PathHandlerConfig{})
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

func TestGeneric_areKeysApplied_true(t *testing.T) {
	client := &vault.MockClient{}

	gh, err := NewGenericHandler(client, PathHandlerConfig{})
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

func TestGeneric_areKeysApplied_false(t *testing.T) {
	client := &vault.MockClient{}

	gh, err := NewGenericHandler(client, PathHandlerConfig{})
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

func TestConvertAuthConfig(t *testing.T) {
	in := vaultApi.AuthConfigInput{}
	_, err := ConvertAuthConfig(in)
	if err != nil {
		log.Fatal(err)
	}
}

// Test that TTLs are converted properly
func TestConvertAuthConfigConvertsDefaultLeaseTTL(t *testing.T) {
	expected := 70
	in := vaultApi.AuthConfigInput{
		DefaultLeaseTTL: "1m10s",
	}
	out, err := ConvertAuthConfig(in)
	if err != nil {
		log.Fatal(err)
	}
	if out.DefaultLeaseTTL != expected {
		log.Fatalf("Wrong DefaultLeastTTL value %d, expected %d", out.DefaultLeaseTTL, expected)
	}
}

func TestConvertAuthConfigConvertsMaxLeaseTTL(t *testing.T) {
	expected := 70
	in := vaultApi.AuthConfigInput{
		MaxLeaseTTL: "1m10s",
	}
	out, err := ConvertAuthConfig(in)
	if err != nil {
		log.Fatal(err)
	}
	if out.MaxLeaseTTL != expected {
		log.Fatalf("Wrong MaxLeastTTL value %d, expected %d", out.MaxLeaseTTL, expected)
	}
}

func TestIsTtlEquivalent(t *testing.T) {
	tests := []struct {
		name string
		ttlA interface{}
		ttlB interface{}
		expected bool
	}{
		{name: "strings", ttlA: "1m", ttlB: "1m", expected: true},
		{name: "ints", ttlA: 60, ttlB: 60, expected: true},
		{name: "string + int", ttlA: "1m", ttlB: 60, expected: true},

		{name: "unequal ints", ttlA: 10, ttlB: 20, expected: false},
		{name: "unequal strings", ttlA: "1m", ttlB: "2m", expected: false},
		{name: "unequal strings + int", ttlA: "1m", ttlB: 120, expected: false},

		{name: "json.Number + string", ttlA: json.Number("60"), ttlB: "1m", expected: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rv := IsTtlEquivalent(test.ttlA, test.ttlB)
			if rv != test.expected {
				log.Fatalf("Test case %s failed. Expected %v, got %v. ttlA: %s, ttlB: %s",
					test.name, test.expected, rv, test.ttlA, test.ttlB)
			}

		})
	}
}

func TestIsSliceEquivalent(t *testing.T) {
	tests := []struct {
		name string
		valueA interface{}
		valueB interface{}
		expected bool
	}{
		{ name: "equal str", valueA: "foo", valueB: "foo", expected: true },
		{ name: "equal str arr", valueA: []string{"foo"}, valueB: []string{"foo"}, expected: true },
		{ name: "equal str + array", valueA: "foo", valueB: []string{"foo"}, expected: true },

		{ name: "unequal str", valueA: "foo", valueB: "bar", expected: false },
		{ name: "unequal str arr", valueA: []string{"foo"}, valueB: []string{"bar"}, expected: false },
		{ name: "unequal str + str array", valueA: "foo", valueB: []string{"bar"}, expected: false },

		{ name: "equal int arr", valueA: []int{99}, valueB: []int{99}, expected: true },
		{ name: "unequal int arr", valueA: []int{99}, valueB: []int{1}, expected: false },
		{ name: "unequal int + int arr", valueA: []int{99}, valueB: 1, expected: false },

		{ name: "equal string + interface", valueA: "foo", valueB: []interface{}{"foo"}, expected: true },
		{ name: "unequal string + interface", valueA: "foo", valueB: []interface{}{"bar"}, expected: false },

		{ name: "empty interfaces", valueA: []interface{}{}, valueB: []interface{}{}, expected: true },
		{ name: "equal interfaces with values", valueA: []interface{}{"foo"}, valueB: []interface{}{"foo"}, expected: true },
		{ name: "unequal interfaces with values", valueA: []interface{}{"foo"}, valueB: []interface{}{"bar"}, expected: false },
		{ name: "unequal interfaces with str int", valueA: []interface{}{"foo"}, valueB: []interface{}{0}, expected: false },
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rv := isSliceEquivalent(test.valueA, test.valueB)
			if rv != test.expected {
				log.Fatalf("Test case %q failed. A: %+v, B: %+v. Expected %+v",
					test.name, test.valueA, test.valueB, test.expected)
			}
		})
	}
}

