package cvault

import (
	"context"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

func (mc *MockVaultClient) CreateUserPass(ctx context.Context, username string, request schema.UserpassWriteUserRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return &vault.Response[map[string]interface{}]{}, nil
}

func (mc *MockVaultClient) ListUserPass(ctx context.Context, options ...vault.RequestOption) (*vault.Response[schema.StandardListResponse], error) {
	return &vault.Response[schema.StandardListResponse]{
		Data: schema.StandardListResponse{
			Keys: []string{"user1", "user2"},
		},
	}, nil
}

func (mc *MockVaultClient) DeleteUserPass(ctx context.Context, username string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return &vault.Response[map[string]interface{}]{}, nil
}
