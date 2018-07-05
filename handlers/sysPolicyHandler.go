package handlers

import (
	"os"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"github.com/starlingbank/vaultsmith/vaultClient"
	"encoding/json"
)

/*
	SysPolicyHandler handles the creation/enabling of auth methods and policies, described in the
	configuration under sys
 */

type SysPolicyHandler struct {
	BasePathHandler
	client					vaultClient.VaultsmithClient
	rootPath				string
	livePolicyList			[]string
	configuredPolicyList	[]string
}

type SysPolicy struct {
	Name	string
	Policy	string `json:"policy"`
}

func NewSysPolicyHandler(c vaultClient.VaultsmithClient, rootPath string) (*SysPolicyHandler, error) {
	// Build a map of currently active auth methods, so walkFile() can reference it
	livePolicyList, err := c.ListPolicies()
	if err != nil {
		return &SysPolicyHandler{}, err
	}

	return &SysPolicyHandler{
		client:              	c,
		rootPath:            	rootPath,
		livePolicyList:      	livePolicyList,
		configuredPolicyList:	[]string{},
	}, nil
}

func (sh *SysPolicyHandler) walkFile(path string, f os.FileInfo, err error) error {
	if f == nil {
		return fmt.Errorf("got nil FileInfo for %s, error: '%s'", path, err.Error())
	}
	if err != nil {
		return fmt.Errorf("error reading %s: %s", path, err)
	}
	// not doing anything with dirs
	if f.IsDir() {
		return nil
	}

	log.Printf("Applying %s\n", path)
	fileContents, err := sh.readFile(path)
	if err != nil {
		return err
	}

	_, file := filepath.Split(path)
	var policy SysPolicy
	err = json.Unmarshal([]byte(fileContents), &policy)
	if err != nil {
		return fmt.Errorf("failed to parse json in %s: %s", path, err)
	}
	policy.Name = file

	err = sh.EnsurePolicy(policy)
	if err != nil {
		return fmt.Errorf("failed to apply policy from %s: %s", path, err)
	}

	return nil
}

func (sh *SysPolicyHandler) PutPoliciesFromDir(path string) error {
	err := filepath.Walk(path, sh.walkFile)
	if err != nil {
		return err
	}
	_, err = sh.RemoveUndeclaredPolicies()
	return err
}

func (sh *SysPolicyHandler) EnsurePolicy(policy SysPolicy) error {
	sh.configuredPolicyList = append(sh.configuredPolicyList, policy.Name)
	applied, err := sh.isPolicyApplied(policy)
	if err != nil {
		return err
	}
	if applied {
		log.Printf("Policy %s already applied", policy.Name)
		return nil
	}
	return sh.client.PutPolicy(policy.Name, policy.Policy)
}

func(sh *SysPolicyHandler) RemoveUndeclaredPolicies() (deleted []string, err error) {
	for _, liveName := range sh.livePolicyList {
		found := false
		for _, configuredName := range sh.configuredPolicyList {
			if liveName == configuredName {
				found = true
				break
			}
		}

		if ! found {
			log.Printf("Deleting policy %s", liveName)
			sh.client.DeletePolicy(liveName)
			deleted = append(deleted, liveName)
		}
	}
	return deleted, nil
}

// true if the policy exists on the server
func (sh *SysPolicyHandler) policyExists(policy SysPolicy) bool {
	//log.Printf("policy.Name: %s, policy list: %+v", policy.Name, sh.livePolicyList)
	for _, p := range sh.livePolicyList {
		if p == policy.Name	{
			return true
		}
	}

	return false
}

// true if the policy is applied on the server
func (sh *SysPolicyHandler) isPolicyApplied(policy SysPolicy) (bool, error) {
	if ! sh.policyExists(policy) {
		return false, nil
	}

	remotePolicy, err := sh.client.GetPolicy(policy.Name)
	if err != nil {
		return false, nil
	}
	log.Printf("%+v", remotePolicy)

	if reflect.DeepEqual(policy.Policy, remotePolicy) {
		return true, nil
	} else {
		return false, nil
	}
}
