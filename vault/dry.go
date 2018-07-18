package vault

import (
	vaultApi "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

type DryClient struct {
	Client
}

// Override any methods that write, so we can only perform reads
func (c *DryClient) EnableAuth(path string, options *vaultApi.EnableAuthOptions) error {
	log.WithFields(log.Fields{
		"action":  "EnableAuth",
		"options": options,
		"path":    path,
	}).WithFields(log.Fields{"foo": "bar"}).Debug()
	return nil
}

func (c *DryClient) DisableAuth(path string) error {
	log.WithFields(log.Fields{
		"action": "DisableAuth",
		"path":   path,
	}).Debug("No action performed")
	return nil
}

func (c *DryClient) PutPolicy(name string, data string) error {
	c.log.WithFields(log.Fields{
		"action": "PutPolicy",
		"name":   name,
		"data":   data,
	}).Debug("No action performed")
	return nil
}

func (c *DryClient) DeletePolicy(name string) error {
	c.log.WithFields(log.Fields{
		"action": "DeletePolicy",
		"name":   name,
	}).Debug("No action performed")
	return nil
}

func (c *DryClient) Read(path string) (*vaultApi.Secret, error) {
	return c.Client.client.Logical().Read(path)
}

func (c *DryClient) Write(path string, data map[string]interface{}) (*vaultApi.Secret, error) {
	c.log.WithFields(log.Fields{
		"action": "Write",
		"path":   path,
		"data":   data,
	}).Debug("No action performed")
	return &vaultApi.Secret{}, nil
}

func (c *DryClient) List(path string) (*vaultApi.Secret, error) {
	return c.Client.client.Logical().List(path)
}
