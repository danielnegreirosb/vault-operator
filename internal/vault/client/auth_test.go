package cvault

import (
	"context"
	"testing"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

func (vc *MockVaultClient) AuthEnableAuthMethod(ctx context.Context, path string, request schema.AuthEnableMethodRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return nil, nil
}

func (vc *MockVaultClient) AuthDisableAuthMethod(ctx context.Context, path string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return nil, nil
}

func (vc *MockVaultClient) ListAuthMethods(ctx context.Context, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return &vault.Response[map[string]interface{}]{}, nil
}


func TestAuthEnabling(t *testing.T) {
	client := &MockVaultClient{}
	authOp := NewAuthOperator(client)

	err := authOp.EnableAuthMethod("my-auth-path", "kubernetes", "token")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestAuthDisabling(t *testing.T) {
	client := &MockVaultClient{}
	authOp := NewAuthOperator(client)

	err := authOp.DisableAuthMethod("my-auth-path", "token")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}