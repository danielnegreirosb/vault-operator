package cvault

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVaultClient(t *testing.T) {

	clientInterface, err := GetClient("https://something")
	require.NoError(t, err)
	client, ok := clientInterface.(*VaultClient)
	require.True(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, "https://something", client.Configuration().Address)
	assert.Equal(t, 60*time.Second, client.Configuration().RequestTimeout)

	clientInterface, err = GetClient("https://something2", WithTimeout(7))
	require.NoError(t, err)
	client, ok = clientInterface.(*VaultClient)
	require.True(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, "https://something2", client.Configuration().Address)
	assert.Equal(t, 7*time.Second, client.Configuration().RequestTimeout)
}
