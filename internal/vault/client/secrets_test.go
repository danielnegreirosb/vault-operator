package cvault

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/stretchr/testify/assert"
)

func TestSecretCreation(t *testing.T) {

	testCases := []struct {
		Name                 string
		secretExists         bool
		expectedCreationCall int
		secretsRandomError   bool
		expectedErr          bool
	}{
		{
			Name:                 "Secret exists",
			secretExists:         true,
			expectedCreationCall: 0,
			secretsRandomError:   false,
			expectedErr:          false,
		},
		{
			Name:                 "Secret does not exist",
			secretExists:         false,
			expectedCreationCall: 1,
			secretsRandomError:   false,
			expectedErr:          false,
		},
		{
			Name:                 "Secret Random Error",
			secretExists:         false,
			expectedCreationCall: 0,
			secretsRandomError:   true,
			expectedErr:          false,
		},
	}

	ctx := context.Background()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			client := &MockVaultClient{secretExists: testCase.secretExists, secretRandomError: testCase.secretsRandomError}
			secretsOp := NewSecretOperator(client)

			err := secretsOp.CreateOrUpdateKvV2Secret(ctx, "", "", "", "", nil)
			if testCase.expectedErr {
				assert.Error(t, err)
			}
			assert.Equal(t, client.secretCreationInvoked, testCase.expectedCreationCall)
		})
	}
}

// func TestRandomizeNested(t *testing.T) {
// 	in := map[string]interface{}{
// 		"a": "{{random}}",
// 		"b": map[string]interface{}{
// 			"c": "{{random}}",
// 			"d": "fixed",
// 		},
// 	}

// 	out := randomize(in, 8)

// 	va, ok := out["a"].(string)
// 	require.True(t, ok)
// 	assert.Equal(t, 8, len(va))
// 	assert.NotEqual(t, "{{random}}", va)

// 	vb, ok := out["b"].(map[string]interface{})
// 	require.True(t, ok)
// 	vc, ok := vb["c"].(string)
// 	require.True(t, ok)
// 	assert.Equal(t, 8, len(vc))
// 	assert.NotEqual(t, "{{random}}", vc)
// 	assert.Equal(t, "fixed", vb["d"])
// }

func TestGenerateRandomStringLengthAndChars(t *testing.T) {
	s := generateRandomString(16)
	assert.Equal(t, 16, len(s))
	// basic sanity: characters should come from the expected charset
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := 0; i < len(s); i++ {
		found := false
		for j := 0; j < len(charset); j++ {
			if s[i] == charset[j] {
				found = true
				break
			}
		}
		assert.True(t, found, "character not in charset")
	}
}

type MockVaultClient struct {

	// input
	secretExists      bool
	secretRandomError bool

	// output
	secretCreationInvoked int
	policyCount           int
}

var _ VaultClientI = (*MockVaultClient)(nil)

func (vc *MockVaultClient) ReadInitializationStatus(ctx context.Context, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return nil, nil
}

func (vc *MockVaultClient) Initialize(ctx context.Context, request schema.InitializeRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return nil, nil
}

func (vc *MockVaultClient) SealStatus(ctx context.Context, options ...vault.RequestOption) (*vault.Response[schema.SealStatusResponse], error) {
	return nil, nil
}

func (vc *MockVaultClient) Unseal(ctx context.Context, request schema.UnsealRequest, options ...vault.RequestOption) (*vault.Response[schema.UnsealResponse], error) {
	return nil, nil
}

func (vc *MockVaultClient) MountsEnableSecretsEngine(ctx context.Context, path string, request schema.MountsEnableSecretsEngineRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return nil, nil
}

func (vc *MockVaultClient) ReadHealthStatus(ctx context.Context, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return nil, nil
}

func (vc *MockVaultClient) KvV2Read(ctx context.Context, path string, options ...vault.RequestOption) (*vault.Response[schema.KvV2ReadResponse], error) {
	if vc.secretRandomError {
		return nil, errors.New("Random error")
	}

	if vc.secretExists {
		return &vault.Response[schema.KvV2ReadResponse]{}, nil
	}
	vc.secretCreationInvoked = 1
	return nil, errors.New("404")
}

func (vc *MockVaultClient) KvV2Write(ctx context.Context, path string, request schema.KvV2WriteRequest, options ...vault.RequestOption) (*vault.Response[schema.KvV2WriteResponse], error) {
	return nil, nil
}

func (vc *MockVaultClient) KvV2Delete(ctx context.Context, path string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return nil, nil
}
