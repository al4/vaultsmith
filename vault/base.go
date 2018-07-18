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
	readMethods
	writeMethods
	Authenticate(string) error
}

type readMethods interface {
	GetPolicy(name string) (string, error)
	List(path string) (*vaultApi.Secret, error)
	ListAuth() (map[string]*vaultApi.AuthMount, error)
	ListPolicies() ([]string, error)
	Read(path string) (*vaultApi.Secret, error)
}

type writeMethods interface {
	Delete(path string) (*vaultApi.Secret, error)
	DeletePolicy(name string) error
	DisableAuth(string) error
	EnableAuth(path string, options *vaultApi.EnableAuthOptions) error
	PutPolicy(string, string) error
	Write(path string, data map[string]interface{}) (*vaultApi.Secret, error)
}

type BaseClient struct {
	readMethods
	writeMethods
	client  *vaultApi.Client
	handler *credAws.CLIHandler
	logger  *log.Entry
}

func NewVaultClient(readonly bool) (c Vault, err error) {
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
	logger := log.WithFields(log.Fields{"readonly": readonly})
	var writer writeMethods
	if readonly {
		writer = &dryClient{
			logger: logger,
		}
	} else {
		writer = &writeClient{
			logger: logger,
			client: vaultApiClient,
		}
	}
	return &BaseClient{
		writeMethods: writer,
		client:       vaultApiClient,
		handler:      &credAws.CLIHandler{},
		logger:       logger,
	}, nil

}

func (c *BaseClient) Authenticate(role string) error {
	if c.client.Token() != "" {
		// Already authenticated. Supposedly.
		c.logger.Debugf("Already authenticated by environment variable")
		return nil
	}

	secret, err := c.handler.Auth(c.client, map[string]string{"role": role})
	if err != nil {
		c.logger.Errorf("Auth error: %s", err)
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

// Only read methods should be in the base client
func (c *BaseClient) Read(path string) (*vaultApi.Secret, error) {
	return c.client.Logical().Read(path)
}

func (c *BaseClient) List(path string) (*vaultApi.Secret, error) {
	return c.client.Logical().List(path)
}

func (c *BaseClient) ListAuth() (map[string]*vaultApi.AuthMount, error) {
	return c.client.Sys().ListAuth()
}

func (c *BaseClient) GetPolicy(name string) (string, error) {
	return c.client.Sys().GetPolicy(name)
}

func (c *BaseClient) ListPolicies() ([]string, error) {
	return c.client.Sys().ListPolicies()
}
