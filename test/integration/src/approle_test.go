package src

import (
	"context"
	"log"
	"os"
	"testing"

	cvault "github.com/danielnegreiros/vault-operator/internal/vault/client"
	"github.com/stretchr/testify/assert"
)

const appRoleName = "my-role"

func TestAppRoleCreate(t *testing.T) {
	for range 3 {
		client, err := cvault.GetClient(vaultAddress, cvault.WithTimeout(5))
		assert.NoError(t, err)
		op := cvault.NewAppRoleOperator(client, vaultAddress)

		err = op.CreateorUpdateAppRole(context.Background(), "approle", appRoleName, 3700, []string{"default"}, os.Getenv("VAULT_TOKEN"))
		assert.NoError(t, err)
	}
}

func TestGetAppRole(t *testing.T) {
	client, err := cvault.GetClient(vaultAddress, cvault.WithTimeout(5))
	assert.NoError(t, err)

	op := cvault.NewAppRoleOperator(client, vaultAddress)
	created, err := op.IsAppRoleCreated(context.Background(), "approle", appRoleName, os.Getenv("VAULT_TOKEN"))
	assert.NoError(t, err)
	assert.True(t, created)
}

func TestDeleteAppRole(t *testing.T) {
	client, err := cvault.GetClient(vaultAddress, cvault.WithTimeout(5))
	assert.NoError(t, err)
	op := cvault.NewAppRoleOperator(client, vaultAddress)
	err = op.DeleteAppRole(context.Background(), "approle", appRoleName, os.Getenv("VAULT_TOKEN"))
	assert.NoError(t, err)
}

func TestGenerateAppRoleSecretID(t *testing.T) {
	client, err := cvault.GetClient(vaultAddress, cvault.WithTimeout(5))
	assert.NoError(t, err)
	op := cvault.NewAppRoleOperator(client, vaultAddress)

	secretId, err := op.GenerateAppRoleSecretID(context.Background(), appRoleName, "approle", os.Getenv("VAULT_TOKEN"))
	assert.NoError(t, err)
	log.Printf("Generated SecretID: %s", secretId)

}