package cvault

import (
	"context"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	vapi "github.com/hashicorp/vault/api"
)

func (mc *MockVaultClient) CreateAppRoleService(ctx context.Context, roleName string, secretIDTTL string, policies []string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return &vault.Response[map[string]interface{}]{}, nil
}

func (mc *MockVaultClient) GetAppRole(ctx context.Context, roleName string, options ...vault.RequestOption) (*vault.Response[schema.AppRoleReadRoleIdResponse], error) {
	return &vault.Response[schema.AppRoleReadRoleIdResponse]{}, nil
}

func (mc *MockVaultClient) DeleteAppRole(ctx context.Context, roleName string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return &vault.Response[map[string]interface{}]{}, nil
}

func (mc *MockVaultClient) WriteAppRoleWithContext(ctx context.Context, path string, roleName string, data map[string]interface{}, ep string, token string) (*vapi.Secret, error) {
	return nil, nil
}
