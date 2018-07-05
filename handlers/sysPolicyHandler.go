package handlers

import (
	"os"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"reflect"
	"github.com/starlingbank/vaultsmith/internal"
)

/*
	SysPolicyHandler handles the creation/enabling of auth methods and policies, described in the
	configuration under sys
 */

type SysPolicyHandler struct {
	BasePathHandler
	client              internal.VaultsmithClient
	rootPath            string
	livePolicyList      []string
	configuredPolicyMap map[string]*string
}

func NewSysPolicyHandler(c internal.VaultsmithClient, rootPath string) (*SysPolicyHandler, error) {
	// Build a map of currently active auth methods, so walkFile() can reference it
	livePolicyList, err := c.ListPolicies()
	if err != nil {
		return &SysPolicyHandler{}, err
	}

	// Create a mapping of configured auth methods, which we append to as we go,
	// so we can disable those that are missing at the end
	configuredPolicyMap := make(map[string]*string)

	return &SysPolicyHandler{
		client:              c,
		rootPath:            rootPath,
		livePolicyList:      livePolicyList,
		configuredPolicyMap: configuredPolicyMap,
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

	dir, file := filepath.Split(path)
	policyPath := strings.Join(strings.Split(dir, "/")[1:], "/")
	//fmt.Printf("path: %s, file: %s\n", policyPath, file)
	if ! strings.HasPrefix(policyPath, "sys/auth") {
		log.Printf("File %s can not be handled yet\n", path)
		return nil
	}

	log.Printf("Applying %s\n", path)
	fileContents, err := sh.readFile(path)
	if err != nil {
		return err
	}

	err = sh.EnsurePolicy(strings.Split(file, ".")[0], fileContents)
	if err != nil {
		return fmt.Errorf("error while ensuring auth for path %s: %s", path, err)
	}

	return nil
}

func (sh *SysPolicyHandler) PutPoliciesFromDir(path string) error {
	err := filepath.Walk(path, sh.walkFile)
	if err != nil {
		return err
	}
	return sh.RemoveUndeclaredPolicies()
}

func (sh *SysPolicyHandler) EnsurePolicy(path string, data string) error {
	// TODO does not check if policy exists
	path = path + "/" // vault appends a slash to paths
	err := sh.client.PutPolicy(path, data)
	if err != nil {
		return fmt.Errorf("could not put policy from %s: %s", path, err)
	}
	return nil
}

func(sh *SysPolicyHandler) RemoveUndeclaredPolicies() error {
	// delete entries not in configured list
	// TODO finish me
	return nil
}

// return true if the localConfig is reflected in remoteConfig, else false
func (sh *SysPolicyHandler) isPolicyApplied(localPolicy string, remotePolicy string) (error, bool) {
	// TODO this will probably not work
	if reflect.DeepEqual(localPolicy, remotePolicy) {
		return nil, true
	} else {
		return nil, false
	}
}
