package path_handlers

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/starlingbank/vaultsmith/document"
	"github.com/starlingbank/vaultsmith/vault"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

/*
	SysPolicy handles the creation/enabling of auth methods and policies, described in the
	configuration under sys

	Unlike SysAuthHandler, it supports templating
*/

// fixed policies that should not be deleted from vault under any circumstances
var fixedPolicies = map[string]bool{
	"root":    true,
	"default": true,
}

type SysPolicy struct {
	BaseHandler
	livePolicyList       []string
	configuredPolicyList []string
}

type policy struct {
	Name       string
	Policy     string `json:"policy"`
	SourceFile string // only for logging
}

func NewSysPolicyHandler(client vault.Vault, config PathHandlerConfig) (*SysPolicy, error) {
	// Build a map of currently active auth methods, so walkFile() can reference it
	livePolicyList, err := client.ListPolicies()
	if err != nil {
		return &SysPolicy{}, fmt.Errorf("error listing policies: %s", err)
	}

	return &SysPolicy{
		BaseHandler: BaseHandler{
			name:   "SysPolicy",
			client: client,
			config: config,
			log: log.WithFields(log.Fields{
				"handler": "SysPolicy",
			}),
		},
		livePolicyList:       livePolicyList,
		configuredPolicyList: []string{},
	}, nil
}

func (sh *SysPolicy) walkFile(path string, f os.FileInfo, err error) error {
	if f == nil {
		sh.log.Infof("%q does not exist, skipping handler. Error was %q", path, err.Error())
		return nil
	}
	if err != nil {
		return fmt.Errorf("error finding %s: %s", path, err)
	}
	// not doing anything with dirs
	if f.IsDir() {
		return nil
	}

	tp, err := document.GenerateTemplateParams(sh.config.TemplateFile, sh.config.TemplateOverrides)
	if err != nil {
		return fmt.Errorf("could not generate template parameters: %s", err)
	}

	content, err := document.Read(path)
	if err != nil {
		return fmt.Errorf("error reading %s: %s", path, err)
	}
	td := &document.Template{
		FileName: strings.TrimSuffix(f.Name(), filepath.Ext(f.Name())),
		Content:  content,
		Params:   tp,
	}

	templatedDocs, err := td.Render()
	if err != nil {
		return fmt.Errorf("failed to render document %q: %s", path, err)
	}

	apiPath, err := apiPath(sh.config.DocumentPath, path)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(apiPath, "sys/policy") {
		return fmt.Errorf("found file without sys/policy prefix: %s", apiPath)
	}
	for _, td := range templatedDocs {
		policy := policy{
			Name:       td.Name,
			SourceFile: f.Name(),
		}
		err = json.Unmarshal([]byte(td.Content), &policy)
		if err != nil {
			return fmt.Errorf("failed to parse json from %s: %s", path, err)
		}

		err = sh.EnsurePolicy(policy)
		if err != nil {
			return fmt.Errorf("failed to apply policy at %s: %s", apiPath, err)
		}
	}

	return nil
}

func (sh *SysPolicy) PutPoliciesFromDir(path string) error {
	err := filepath.Walk(path, sh.walkFile)
	if err != nil {
		return err
	}
	_, err = sh.RemoveUndeclaredPolicies()
	return err
}

func (sh *SysPolicy) EnsurePolicy(policy policy) error {
	logger := sh.log.WithFields(log.Fields{
		"name":       policy.Name,
		"sourceFile": policy.SourceFile,
	})

	sh.configuredPolicyList = append(sh.configuredPolicyList, policy.Name)
	applied, err := sh.isPolicyApplied(policy)
	if err != nil {
		return err
	}
	if applied {
		logger.Debugf("Policy already applied")
		return nil
	}
	logger.Info("Applying policy")
	return sh.client.PutPolicy(policy.Name, policy.Policy)
}

func (sh *SysPolicy) RemoveUndeclaredPolicies() (deleted []string, err error) {
	// only real reason to track the deleted policies is for testing as logs inform user
	for _, liveName := range sh.livePolicyList {
		if fixedPolicies[liveName] {
			// never want to delete default or root
			continue
		}

		// look for the policy in the configured list
		found := false
		for _, configuredName := range sh.configuredPolicyList {
			if liveName == configuredName {
				found = true // it's declared and should stay
				break
			}
		}

		if !found {
			// not declared, delete
			sh.log.WithFields(log.Fields{"policy": liveName}).Infof("Deleting policy")
			sh.client.DeletePolicy(liveName)
			deleted = append(deleted, liveName)
		}
	}
	return deleted, nil
}

// true if the policy exists on the server
func (sh *SysPolicy) policyExists(policy policy) bool {
	//sh.log.Debugf("policy.Name: %s, policy list: %+v", policy.Name, sh.livePolicyList)
	for _, p := range sh.livePolicyList {
		if p == policy.Name {
			return true
		}
	}

	return false
}

// true if the policy is applied on the server
func (sh *SysPolicy) isPolicyApplied(policy policy) (bool, error) {
	if !sh.policyExists(policy) {
		return false, nil
	}

	remotePolicy, err := sh.client.GetPolicy(policy.Name)
	if err != nil {
		return false, nil
	}

	// TODO Need a proper HCL parser here, testing strings is error prone
	if reflect.DeepEqual(policy.Policy, remotePolicy) {
		return true, nil
	} else {
		log.Debugf("Policy not equal (local != remote): \n%+v\n!=\n%+v\n", policy.Policy, remotePolicy)
		return false, nil
	}
}

func (sh *SysPolicy) Order() int {
	return sh.order
}
