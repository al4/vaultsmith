package handlers

import (
	"github.com/starlingbank/vaultsmith/vaultClient"
	"log"
	"testing"
)


func TestPolicyExists(t *testing.T) {
	// Not terribly testable as it doesn't return anything we can assert against
	client := &vaultClient.MockVaultsmithClient{}
	sph, err := NewSysPolicyHandler(client, "")
	if err != nil {
		log.Fatalf("Failed to create SysAuthHandler: %s", err)
	}

	p := SysPolicy{
		Name: "testName",
		Policy: "testPolicy",
	}
	sph.livePolicyList = []string{"testName"}
	r := sph.policyExists(p)
	if err != nil {
		log.Fatalf("Error calling policyExists: %s", err)
	}
	if ! r {
		log.Fatalf("Policy does not exist, yet it does (expected true)")
	}
}

func TestPolicyExistsFalse(t *testing.T) {
	client := &vaultClient.MockVaultsmithClient{}
	sph, err := NewSysPolicyHandler(client, "")
	if err != nil {
		log.Fatalf("Failed to create SysAuthHandler: %s", err)
	}

	p := SysPolicy{
		Name: "testName",
		Policy: "testPolicy",
	}
	sph.livePolicyList = []string{}
	r := sph.policyExists(p)
	if err != nil {
		log.Fatalf("Error calling policyExists: %s", err)
	}
	if r {
		log.Fatalf("Policy exists, yet it does not (expected false)")
	}
}

type getPolicyClient struct {
	vaultClient.MockVaultsmithClient
	returnValue string
}

func (c *getPolicyClient) GetPolicy(name string) (string, error) {
	return c.returnValue, nil
}

// isPolicyApplied should return true when the policy is present and the content matches
func TestIsPolicyApplied(t *testing.T) {
	client := &getPolicyClient{
		returnValue: "testPolicy",
	}
	sph, err := NewSysPolicyHandler(client, "")
	if err != nil {
		log.Fatalf("Failed to create SysAuthHandler: %s", err)
	}

	p := SysPolicy{
		Name: "testName",
		Policy: "testPolicy",
	}
	sph.livePolicyList = []string{"testName"}
	rv, err := sph.isPolicyApplied(p)
	if err != nil {
		log.Fatalf("Error calling isPolicyApplied: %s", err)
	}
	if ! rv {
		log.Fatalf("isPolicyApplied returns false, should be true in this case")
	}
}

// isPolicyApplied should return false when policy is present but content differs
func TestIsPolicyApplied_PresentButDifferent(t *testing.T) {
	client := &getPolicyClient{
		returnValue: "testPolicy",
	}
	sph, err := NewSysPolicyHandler(client, "")
	if err != nil {
		log.Fatalf("Failed to create SysAuthHandler: %s", err)
	}

	p := SysPolicy{
		Name: "testName",
		Policy: "this content is different",
	}
	sph.livePolicyList = []string{"testName"}
	rv, err := sph.isPolicyApplied(p)
	if err != nil {
		log.Fatalf("Error calling isPolicyApplied: %s", err)
	}
	if rv {
		log.Fatalf("isPolicyApplied returns true, should be false in this case")
	}
}
