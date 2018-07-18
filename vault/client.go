package vault

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"

	"crypto/tls"
	vaultApi "github.com/hashicorp/vault/api"
	credAws "github.com/hashicorp/vault/builtin/credential/aws"
)

/*
With the exception of Authenticate, most functions in this file are simple pass-through calls
to the vault API, which don't do anything special. They should however, be idempotent, and thus
not return an error if that error indicates that the operation has already been done, e.g.
"already exists" type errors.

If there is a possibility that the configuration might be different, they should delete and then
put.
*/

// Vault is an abstraction of hashicorp's vault api client
type Vault interface {
	Authenticate(string) error
	DeletePolicy(name string) error
	DisableAuth(string) error
	EnableAuth(path string, options *vaultApi.EnableAuthOptions) error
	GetPolicy(name string) (string, error)
	ListAuth() (map[string]*vaultApi.AuthMount, error)
	ListPolicies() ([]string, error)
	PutPolicy(string, string) error
	Read(path string) (*vaultApi.Secret, error)
	Write(path string, data map[string]interface{}) (*vaultApi.Secret, error)
	List(path string) (*vaultApi.Secret, error)
}

type Client struct {
	client  *vaultApi.Client
	handler *credAws.CLIHandler
	log     *log.Entry
}

func NewVaultClient(dry bool) (c Vault, err error) {
	config := vaultApi.Config{
		HttpClient: &http.Client{
			Transport: &http.Transport{
				// lack of TLSClientConfig can cause SIGSEGV on config.ReadEnvironment() below
				// when VAULT_SKIP_VERIFY is true
				TLSClientConfig: &tls.Config{},
			},
		},
	}

	err = config.ReadEnvironment()
	if err != nil {
		return c, err
	}

	vaultApiClient, err := vaultApi.NewClient(&config)
	if err != nil {
		return c, err
	}

	client := Client{
		log:     log.WithFields(log.Fields{"dry": dry}),
		client:  vaultApiClient,
		handler: &credAws.CLIHandler{},
	}
	if dry {
		c = &DryClient{
			Client: client,
		}
	} else {
		return &client, nil
	}

	return c, nil
}

func (c *Client) Authenticate(role string) error {
	if c.client.Token() != "" {
		// Already authenticated. Supposedly.
		c.log.Debugf("Already authenticated by environment variable")
		return nil
	}

	secret, err := c.handler.Auth(c.client, map[string]string{"role": role})
	if err != nil {
		c.log.Errorf("Auth error: %s", err)
		return err
	}

	if secret == nil {
		return errors.New("no secret returned from Vault")
	}

	c.client.SetToken(secret.Auth.ClientToken)

	secret, err = c.client.Auth().Token().LookupSelf()
	if err != nil {
		return errors.New(fmt.Sprintf("no token found in Vault client (%s)", err))
	}

	return nil
}

// Used by sysAuthHandler
func (c *Client) EnableAuth(path string, options *vaultApi.EnableAuthOptions) error {
	c.log.WithFields(log.Fields{
		"action":  "EnableAuth",
		"options": options,
		"path":    path,
	}).Debug()
	return c.client.Sys().EnableAuthWithOptions(path, options)
}

func (c *Client) ListAuth() (map[string]*vaultApi.AuthMount, error) {
	return c.client.Sys().ListAuth()
}

func (c *Client) DisableAuth(path string) error {
	c.log.WithFields(log.Fields{
		"action": "DisableAuth",
		"path":   path,
	}).Debug("Performing action")
	return c.client.Sys().DisableAuth(path)
}

// Used by sysPolicyHandler
func (c *Client) ListPolicies() ([]string, error) {
	return c.client.Sys().ListPolicies()
}

func (c *Client) GetPolicy(name string) (string, error) {
	return c.client.Sys().GetPolicy(name)
}

func (c *Client) PutPolicy(name string, data string) error {
	c.log.WithFields(log.Fields{
		"action": "PutPolicy",
		"name":   name,
		"data":   data,
	}).Debug("Performing action")
	return c.client.Sys().PutPolicy(name, data)
}

func (c *Client) DeletePolicy(name string) error {
	c.log.WithFields(log.Fields{
		"action": "DeletePolicy",
		"name":   name,
	}).Debug("Performing action")
	return c.client.Sys().DeletePolicy(name)
}

// Used by genericHandler
func (c *Client) Read(path string) (*vaultApi.Secret, error) {
	return c.client.Logical().Read(path)
}

func (c *Client) Write(path string, data map[string]interface{}) (*vaultApi.Secret, error) {
	c.log.WithFields(log.Fields{
		"action": "Write",
		"path":   path,
		"data":   data,
	}).Debug("Performing action")
	return c.client.Logical().Write(path, data)
}

func (c *Client) List(path string) (*vaultApi.Secret, error) {
	return c.client.Logical().List(path)
}
