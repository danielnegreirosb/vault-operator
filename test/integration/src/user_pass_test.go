package src

import (
	"os"
	"testing"

	cvault "github.com/danielnegreiros/vault-operator/internal/vault/client"
	"github.com/stretchr/testify/assert"
)

func TestUserPassCreate(t *testing.T) {
	for range 3 {
		client, err := cvault.GetClient(vaultAddress, cvault.WithTimeout(5))
		assert.NoError(t, err)
		op := cvault.NewUserPassOperator(client)

		err = op.CreateUser("userpass", "user2", "changeme", []string{"default"}, os.Getenv("VAULT_TOKEN"))
		assert.NoError(t, err)
	}
}

func TestIsUserCreated(t *testing.T) {
	client, err := cvault.GetClient(vaultAddress, cvault.WithTimeout(5))
	assert.NoError(t, err)

	op := cvault.NewUserPassOperator(client)
	created, err := op.IsUserCreated(client, "userpass_strange", "user2", "hvs.zLqeOcu5HftaAHM4rPqU4WZW")
	assert.NoError(t, err)
	assert.True(t, created)
}

func TestDeleteUserPass(t *testing.T) {
	client, err := cvault.GetClient(vaultAddress, cvault.WithTimeout(5))
	assert.NoError(t, err)
	op := cvault.NewUserPassOperator(client)

	err = op.DeleteUserPass("userpass", "user2", "hvs.zLqeOcu5HftaAHM4rPqU4WZW")
	assert.NoError(t, err)
}
