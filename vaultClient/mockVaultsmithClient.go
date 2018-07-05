package vaultClient

import (
	"fmt"
	"github.com/stretchr/testify/mock"
	vaultApi "github.com/hashicorp/vault/api"
)

type MockVaultsmithClient struct {
	mock.Mock
	ReturnString string
	ReturnError  error
}

func (m *MockVaultsmithClient) Authenticate(role string) error {
	m.Called(role)
	if role == "ConnectionRefused" {
		return fmt.Errorf("dial tcp [::1]:8200: getsockopt: connection refused")
	} else if role == "InvalidRole" {
		return fmt.Errorf("entry for role InvalidRole not found")
	}
	return m.ReturnError
}

func (m *MockVaultsmithClient) DisableAuth(string) error {
	return m.ReturnError
}

func (m *MockVaultsmithClient) EnableAuth(path string, options *vaultApi.EnableAuthOptions) error {
	return m.ReturnError
}

func (m *MockVaultsmithClient) ListAuth() (map[string]*vaultApi.AuthMount, error) {
	rv := make(map[string]*vaultApi.AuthMount)
	return rv, m.ReturnError
}

func (m *MockVaultsmithClient) ListPolicies() ([]string, error) {
	rv := make([]string, 0)
	return rv, m.ReturnError
}

func (m *MockVaultsmithClient) GetPolicy(name string) (string, error) {
	return m.ReturnString, m.ReturnError
}

func (m *MockVaultsmithClient) PutPolicy(name string, data string) error {
	return m.ReturnError
}

func (m *MockVaultsmithClient) DeletePolicy(name string) (error) {
	return m.ReturnError
}
