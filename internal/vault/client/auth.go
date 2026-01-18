package cvault

import (
	"context"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

type AuthOperator struct {
	client VaultClientI
}

func NewAuthOperator(client VaultClientI) *AuthOperator {
	return &AuthOperator{client: client}
}

func (ao *AuthOperator) EnableAuthMethod(path string, method string, token string) error {
	ok, err := ao.IsAuthMethodEnabled(ao.client, path, token)
	if err != nil {
		return err
	}

	if ok {
		return nil
	}

	_, err = ao.client.AuthEnableAuthMethod(
		context.Background(),
		path,
		schema.AuthEnableMethodRequest{
			Type: method,
		},
		vault.WithToken(token),
	)
	return err
}

func (ao *AuthOperator) DisableAuthMethod(path string, token string) error {
	_, err := ao.client.AuthDisableAuthMethod(
		context.Background(),
		path,
		vault.WithToken(token),
	)
	return err
}

func (ao *AuthOperator) IsAuthMethodEnabled(client VaultClientI, path string, token string) (bool, error) {
	resp, err := client.ListAuthMethods(context.Background(), vault.WithToken(token))
	if err != nil {
		return false, err
	}

	if _, ok := resp.Data[path+"/"]; ok {
		return true, nil
	}
	return false, nil
}

func (vc *VaultClient) AuthEnableAuthMethod(ctx context.Context, path string, request schema.AuthEnableMethodRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.System.AuthEnableMethod(ctx, path, request, options...)
}

func (vc *VaultClient) AuthDisableAuthMethod(ctx context.Context, path string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.System.AuthDisableMethod(ctx, path, options...)
}

func (vc *VaultClient) ListAuthMethods(ctx context.Context, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.System.AuthListEnabledMethods(ctx, options...)
}
