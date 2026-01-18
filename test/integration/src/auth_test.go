package src

import (
	"os"
	"testing"

	cvault "github.com/danielnegreiros/vault-operator/internal/vault/client"
	"github.com/stretchr/testify/assert"
)

const (
	vaultAddress = "http://localhost:8200"
)

func TestAuthEnable(t *testing.T) {
	for range 3 {
		client, err := cvault.GetClient(vaultAddress, cvault.WithTimeout(5))
		assert.NoError(t, err)
		op := cvault.NewAuthOperator(client)

		err = op.EnableAuthMethod("kubernetes", "kubernetes", os.Getenv("VAULT_TOKEN"))
		assert.NoError(t, err)
	}
}

func TestIsAuthMethodEnabled(t *testing.T) {
	client, err := cvault.GetClient(vaultAddress, cvault.WithTimeout(5))
	assert.NoError(t, err)

	op := cvault.NewAuthOperator(client)
	enabled, err := op.IsAuthMethodEnabled(client, "kubernetes", os.Getenv("VAULT_TOKEN"))
	assert.NoError(t, err)
	assert.True(t, enabled)
}

func TestAuthDisable(t *testing.T) {
	client, err := cvault.GetClient(vaultAddress, cvault.WithTimeout(5))
	assert.NoError(t, err)
	op := cvault.NewAuthOperator(client)

	err = op.DisableAuthMethod("kubernetes", os.Getenv("VAULT_TOKEN"))
	assert.NoError(t, err)
}
