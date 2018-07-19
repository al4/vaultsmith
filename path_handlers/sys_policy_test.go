package path_handlers

import (
	log "github.com/sirupsen/logrus"
	"github.com/starlingbank/vaultsmith/vault"
	"reflect"
	"testing"
)

func TestSysPolicyHandler_PolicyExists(t *testing.T) {
	// Not terribly testable as it doesn't return anything we can assert against
	client := &vault.MockClient{}
	sph, err := NewSysPolicyHandler(client, PathHandlerConfig{})
	if err != nil {
		t.Errorf("Failed to create SysAuth: %s", err)
	}

	p := policy{
		Name:   "testName",
		Policy: "testPolicy",
	}
	sph.livePolicyList = []string{"testName"}
	r := sph.policyExists(p)
	if err != nil {
		t.Errorf("Error calling policyExists: %s", err)
	}
	if !r {
		t.Errorf("Policy does not exist, yet it does (expected true)")
	}
}

func TestSysPolicyHandler_PolicyExistsFalse(t *testing.T) {
	client := &vault.MockClient{}
	sph, err := NewSysPolicyHandler(client, PathHandlerConfig{})
	if err != nil {
		t.Errorf("Failed to create SysAuth: %s", err)
	}

	p := policy{
		Name:   "testName",
		Policy: "testPolicy",
	}
	sph.livePolicyList = []string{}
	r := sph.policyExists(p)
	if err != nil {
		t.Errorf("Error calling policyExists: %s", err)
	}
	if r {
		t.Errorf("Policy exists, yet it does not (expected false)")
	}
}

// isPolicyApplied should return true when the policy is present and the content matches
func TestSysPolicyHandler_IsPolicyApplied(t *testing.T) {
	client := &vault.MockClient{}
	client.ReturnString = "testPolicy"
	sph, err := NewSysPolicyHandler(client, PathHandlerConfig{})
	if err != nil {
		t.Errorf("Failed to create SysAuth: %s", err)
	}

	p := policy{
		Name:   "testName",
		Policy: "testPolicy",
	}
	sph.livePolicyList = []string{"testName"}
	rv, err := sph.isPolicyApplied(p)
	if err != nil {
		t.Errorf("Error calling isPolicyApplied: %s", err)
	}
	if !rv {
		t.Errorf("isPolicyApplied returns false, should be true in this case")
	}
}

// isPolicyApplied should return false when policy is present but content differs
func TestSysPolicyHandler_IsPolicyApplied_PresentButDifferent(t *testing.T) {
	client := &vault.MockClient{}
	client.ReturnString = "testPolicy"

	sph, err := NewSysPolicyHandler(client, PathHandlerConfig{})
	if err != nil {
		t.Errorf("Failed to create SysAuth: %s", err)
	}

	p := policy{
		Name:   "testName",
		Policy: "this content is different",
	}
	sph.livePolicyList = []string{"testName"}
	rv, err := sph.isPolicyApplied(p)
	if err != nil {
		t.Errorf("Error calling isPolicyApplied: %s", err)
	}
	if rv {
		t.Errorf("isPolicyApplied returns true, should be false in this case")
	}
}

func TestSysPolicyHandler_RemoveUndeclaredPolicies(t *testing.T) {
	sph, err := NewSysPolicyHandler(&vault.MockClient{}, PathHandlerConfig{})
	if err != nil {
		t.Errorf("Failed to create SysAuth: %s", err)
	}

	sph.livePolicyList = []string{"foo", "qux", "bar", "baz", "quux"}
	sph.configuredPolicyList = []string{"baz", "foo", "bar"}

	expected := []string{"qux", "quux"}
	deleted, err := sph.RemoveUndeclaredPolicies()
	if err != nil {
		log.Fatal(err)
	}
	if !reflect.DeepEqual(deleted, expected) {
		t.Errorf("List of deleted policies does not match expected (%+v != %+v)",
			deleted, expected)
	}
}
