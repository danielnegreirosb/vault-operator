package src

import (
	"os"
	"testing"

	cvault "github.com/danielnegreiros/vault-operator/internal/vault/client"
	"github.com/stretchr/testify/assert"
)

func TestMountEnabling(t *testing.T) {
	client, err := cvault.GetClient(vaultAddress, cvault.WithTimeout(5))
	assert.NoError(t, err)
	secretEngOp := cvault.NewSecretEngineOperator(client)

	err = secretEngOp.EnableMount("my-mount-path", "kv", os.Getenv("VAULT_TOKEN"))
	assert.NoError(t, err)
}

func TestMountDisabling(t *testing.T) {
	client, err := cvault.GetClient(vaultAddress, cvault.WithTimeout(5))
	assert.NoError(t, err)
	secretEngOp := cvault.NewSecretEngineOperator(client)

	err = secretEngOp.DisableMount("my-mount-path", os.Getenv("VAULT_TOKEN"))
	assert.NoError(t, err)
}

func TestIsMountEnabled(t *testing.T) {
	client, err := cvault.GetClient(vaultAddress, cvault.WithTimeout(5))
	assert.NoError(t, err)

	secretEngOp := cvault.NewSecretEngineOperator(client)

	testCases := []struct {
		path         string
		expectedBool bool
	}{
		{"my-mount-path/", true},
		{"nonexistent/", false},
	}

	for _, tc := range testCases {
		enabled, err := secretEngOp.IsMountEnabled(tc.path, os.Getenv("VAULT_TOKEN"))
		assert.NoError(t, err)
		assert.Equal(t, tc.expectedBool, enabled)
	}
}
