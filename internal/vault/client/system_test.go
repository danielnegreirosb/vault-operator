package cvault

// import (
// 	"context"
// 	"log"
// 	"os"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// var extractedToken string
// var keys []interface{}

// func initVault(t *testing.T, vclient *VaultOperator) {
// 	data, err := vclient.InitVault(context.Background())
// 	require.NoError(t, err)
// 	var ok bool
// 	extractedToken, ok = data["root_token"].(string)
// 	log.Println(data)
// 	require.True(t, ok)
// 	keys, ok = data["keys"].([]interface{})
// 	require.True(t, ok)
// }

// func unseal(t *testing.T, vclient *VaultOperator){
// 	err := vclient.Unseal(context.Background(), keys)
// 	require.NoError(t, err)
// }

// func TestIsInitialized(t *testing.T) {

// 	require.NotEmpty(t, os.Getenv("VAULT_URL_DOCKER"))

// 	testCases := []struct {
// 		Name        string
// 		Url         string
// 		isInit      bool
// 		expectedErr string
// 	}{
// 		{
// 			Name:        "Testing Is Initialized true",
// 			Url:         os.Getenv("VAULT_URL_DOCKER"),
// 			isInit:      false,
// 			expectedErr: "",
// 		},
// 		{
// 			Name:        "Test Is Initialized false",
// 			Url:         os.Getenv("VAULT_URL_DOCKER"),
// 			isInit:      true,
// 			expectedErr: "",
// 		},
// 		{
// 			Name:        "Test Connection Error",
// 			Url:         "http://10.10.10.10:8200",
// 			isInit:      true,
// 			expectedErr: "deadline",
// 		},
// 	}

// 	for i, testCase := range testCases {
// 		t.Run(testCase.Name, func(t *testing.T) {
// 			vclient, err := GetClient(testCase.Url, WithTimeout(1))
// 			assert.NoError(t, err)

// 			isInit, err := vclient.IsInitialized(context.Background())
// 			if testCase.expectedErr != "" {
// 				assert.ErrorContains(t, err, testCase.expectedErr)
// 			} else {
// 				assert.Equal(t, isInit, testCase.isInit)
// 			}

// 			if i == 0 {
// 				initVault(t, vclient)
// 			}
// 		})
// 	}
// 	log.Println(extractedToken)
// }

// func TestEnableSecretEngine(t *testing.T) {

// 	require.NotEmpty(t, os.Getenv("VAULT_URL_DOCKER"))

// 	testCases := []struct {
// 		Name string
// 		Url  string
// 	}{
// 		{
// 			Name: "Secret Engine is mount",
// 			Url:  os.Getenv("VAULT_URL_DOCKER"),
// 		},
// 	}

// 	ctx := context.Background()
// 	for i, testCase := range testCases {
// 		t.Run(testCase.Name, func(t *testing.T) {
// 			vclient, err := GetClient(testCase.Url, WithTimeout(2))
// 			assert.NoError(t, err)
// 			if i == 0 {
// 				initVault(t, vclient)
// 				unseal(t, vclient)
// 			}

// 			err = vclient.mountKvEnginePath(ctx, "mysecret", "kv-v2", extractedToken)
// 			assert.NoError(t, err)
// 		})
// 	}

// }
