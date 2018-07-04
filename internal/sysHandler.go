package internal

import (
	"os"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"reflect"
	vaultApi "github.com/hashicorp/vault/api"
	"encoding/json"
)

/*
	SysHandler handles the creation/enabling of auth methods and policies, described in the
	configuration under sys
 */

type SysHandler struct {
	BasePathHandler
	client 				VaultsmithClient
	rootPath 			string
	liveAuthMap 		map[string]*vaultApi.AuthMount
	configuredAuthMap 	map[string]*vaultApi.AuthMount
}

func NewSysHandler(c VaultsmithClient, rootPath string) (*SysHandler, error) {
	// Build a map of currently active auth methods, so walkFile() can reference it
	liveAuthMap, err := c.ListAuth()
	if err != nil {
		return &SysHandler{}, err
	}

	// Create a mapping of configured auth methods, which we append to as we go,
	// so we can disable those that are missing at the end
	configuredAuthMap := make(map[string]*vaultApi.AuthMount)

	return &SysHandler{
		client: c,
		rootPath: rootPath,
		liveAuthMap: liveAuthMap,
		configuredAuthMap: configuredAuthMap,
	}, nil
}

func (sh *SysHandler) walkFile(path string, f os.FileInfo, err error) error {
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
	var enableOpts vaultApi.EnableAuthOptions
	err = json.Unmarshal([]byte(fileContents), &enableOpts)
	if err != nil {
		return fmt.Errorf("could not parse json from file %s: %s", path, err)
	}

	err = sh.EnsureAuth(strings.Split(file, ".")[0], enableOpts)
	if err != nil {
		return fmt.Errorf("error while ensuring auth for path %s: %s", path, err)
	}

	return nil
}

func (sh *SysHandler) PutPoliciesFromDir(path string) error {
	err := filepath.Walk(path, sh.walkFile)
	if err != nil {
		return err
	}
	return sh.DisableUnconfiguredAuths()
}

// Ensure that this auth type is enabled and has the correct configuration
func (sh *SysHandler) EnsureAuth(path string, enableOpts vaultApi.EnableAuthOptions) error {
	// we need to convert to AuthConfigOutput in order to compare with existing config
	var enableOptsAuthConfigOutput vaultApi.AuthConfigOutput
	enableOptsAuthConfigOutput, err := ConvertAuthConfig(enableOpts.Config)
	if err != nil {
		return err
	}

	authMount := vaultApi.AuthMount{
		Type:   enableOpts.Type,
		Config: enableOptsAuthConfigOutput,
	}
	sh.configuredAuthMap[path] = &authMount

	path = path + "/" // vault appends a slash to paths
	if liveAuth, ok := sh.liveAuthMap[path]; ok {
		// If this path is present in our live config, we may not need to enable
		err, applied := sh.isConfigApplied(enableOpts.Config, liveAuth.Config)
		if err != nil {
			return fmt.Errorf(
				"could not determine whether configuration for auth mount %s was applied: %s",
				enableOpts.Type, err)
		}
		if applied {
			log.Printf("Configuration for authMount %s already applied\n", enableOpts.Type)
			return nil
		}
	}
	log.Printf("Enabling auth type %s\n", authMount.Type)
	err = sh.client.EnableAuth(path, &enableOpts)
	if err != nil {
		return fmt.Errorf("could not enable auth %s: %s", path, err)
	}
	return nil
}

func(sh *SysHandler) DisableUnconfiguredAuths() error {
	// delete entries not in configured list
	for k, authMount := range sh.liveAuthMap {
		path := strings.Trim(k, "/") // vault appends a slash to paths
		if _, ok := sh.configuredAuthMap[path]; ok {
			continue  // present, do nothing
		} else if authMount.Type == "token" {
			continue  // cannot be disabled, would give http 400 if attempted
		} else {
			log.Printf("Disabling auth type %s\n", authMount.Type)
			return sh.client.DisableAuth(authMount.Type)
		}
	}
	return nil
}

// return true if the localConfig is reflected in remoteConfig, else false
func (sh *SysHandler) isConfigApplied(localConfig vaultApi.AuthConfigInput, remoteConfig vaultApi.AuthConfigOutput) (error, bool) {
	// AuthConfigInput uses different types for TTL, which need to be converted
	converted, err := ConvertAuthConfig(localConfig)
	if err != nil {
		return err, false
	}

	if reflect.DeepEqual(converted, remoteConfig) {
		return nil, true
	} else {
		return nil, false
	}
}
