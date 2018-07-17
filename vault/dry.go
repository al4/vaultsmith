package vault
import (
	vaultApi "github.com/hashicorp/vault/api"
)

type DryClient struct {
	Client
}

// Override any methods that write, so we can only perform reads
func (c *DryClient) EnableAuth(path string, options *vaultApi.EnableAuthOptions) error {
	return nil
}

func (c *DryClient) DisableAuth(path string) error {
	return nil
}

func (c *DryClient) PutPolicy(name string, data string) error {
	return nil
}

func (c *DryClient) DeletePolicy(name string) (error) {
	return nil
}

func (c *DryClient) Read(path string) (*vaultApi.Secret, error) {
	return c.Client.client.Logical().Read(path)
}

func (c *DryClient) Write(path string, data map[string]interface{}) (*vaultApi.Secret, error) {
	return &vaultApi.Secret{}, nil
}

func (c *DryClient) List(path string) (*vaultApi.Secret, error) {
	return c.Client.client.Logical().List(path)
}
