package vault

import (
	"fmt"
	"github.com/stretchr/testify/mock"
	vaultApi "github.com/hashicorp/vault/api"
)

type MockClient struct {
	mock.Mock
	ReturnString string
	ReturnError  error
	ReturnSecret *vaultApi.Secret
}

func (m *MockClient) Authenticate(role string) error {
	m.Called(role)
	if role == "ConnectionRefused" {
		return fmt.Errorf("dial tcp [::1]:8200: getsockopt: connection refused")
	} else if role == "InvalidRole" {
		return fmt.Errorf("entry for role InvalidRole not found")
	}
	return m.ReturnError
}

func (m *MockClient) DisableAuth(string) error {
	return m.ReturnError
}

func (m *MockClient) EnableAuth(path string, options *vaultApi.EnableAuthOptions) error {
	return m.ReturnError
}

func (m *MockClient) ListAuth() (map[string]*vaultApi.AuthMount, error) {
	rv := make(map[string]*vaultApi.AuthMount)
	return rv, m.ReturnError
}

func (m *MockClient) ListPolicies() ([]string, error) {
	rv := make([]string, 0)
	return rv, m.ReturnError
}

func (m *MockClient) GetPolicy(name string) (string, error) {
	return m.ReturnString, m.ReturnError
}

func (m *MockClient) PutPolicy(name string, data string) error {
	return m.ReturnError
}

func (m *MockClient) DeletePolicy(name string) (error) {
	return m.ReturnError
}

func (m *MockClient) Read(path string) (*vaultApi.Secret, error) {
	return m.ReturnSecret, m.ReturnError
}

func (m *MockClient) Write(path string, data map[string]interface{}) (*vaultApi.Secret, error) {
	return m.ReturnSecret, m.ReturnError
}

func (m *MockClient) List(path string) (*vaultApi.Secret, error) {
	return m.ReturnSecret, m.ReturnError
}
