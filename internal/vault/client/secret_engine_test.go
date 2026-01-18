package cvault

import (
	"context"
	"testing"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/stretchr/testify/assert"
)

func (mc *MockVaultClient) MountEnable(ctx context.Context, path string, request schema.MountsEnableSecretsEngineRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return &vault.Response[map[string]interface{}]{}, nil
}

func (mc *MockVaultClient) MountDisable(ctx context.Context, path string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return &vault.Response[map[string]interface{}]{}, nil
}

func (mc *MockVaultClient) ListMounts(ctx context.Context, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	// return fake secret path is mounted in the response
	return &vault.Response[map[string]interface{}]{
		Data: map[string]interface{}{
			"secret/": map[string]interface{}{},
		},
	}, nil
}


func TestMountEnabling(t *testing.T) {
	client := &MockVaultClient{}
	secretEngOp := NewSecretEngineOperator(client)

	err := secretEngOp.EnableMount("my-mount-path", "kv", "token")
	assert.NoError(t, err)
}

func TestMountDisabling(t *testing.T){
	client := &MockVaultClient{}
	secretEngOp := NewSecretEngineOperator(client)

	err := secretEngOp.DisableMount("my-mount-path", "token")
	assert.NoError(t, err)
}

func TestIsMountEnabled(t *testing.T) {
	client := &MockVaultClient{}
	secretEngOp := NewSecretEngineOperator(client)

	testCases := []struct {
		path         string
		expectedBool bool
	}{
		{"secret/", true},
		{"nonexistent/", false},
	}

	for _, tc := range testCases {
		enabled, err := secretEngOp.IsMountEnabled(tc.path, "token")
		assert.NoError(t, err)
		assert.Equal(t, tc.expectedBool, enabled)
	}
}