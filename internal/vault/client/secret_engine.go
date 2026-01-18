package cvault

import (
	"context"
	"strings"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

type SecretEngineOperator struct {
	client VaultClientI
}

func NewSecretEngineOperator(client VaultClientI) *SecretEngineOperator {
	return &SecretEngineOperator{client: client}
}

func (mo *SecretEngineOperator) IsMountEnabled(path string, token string) (bool, error) {
	resp, err := mo.client.ListMounts(context.Background(), vault.WithToken(token))
	if err != nil {
		return false, err
	}

	path = strings.TrimSuffix(path, "/")

	if _, ok := resp.Data[path+"/"]; ok {
		return true, nil
	}
	return false, nil
}

func (mo *SecretEngineOperator) EnableMount(path string, mountType string, token string) error {
	ok, err := mo.IsMountEnabled(path, token)
	if err != nil {
		return err
	}

	if ok {
		return nil
	}

	_, err = mo.client.MountEnable(
		context.Background(),
		path,
		schema.MountsEnableSecretsEngineRequest{
			Type: mountType,
		},
		vault.WithToken(token),
	)
	return err
}

func (mo *SecretEngineOperator) DisableMount(path string, token string) error {
	_, err := mo.client.MountDisable(
		context.Background(),
		path,
		vault.WithToken(token),
	)
	return err
}

func (vc *VaultClient) MountEnable(ctx context.Context, path string, request schema.MountsEnableSecretsEngineRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.System.MountsEnableSecretsEngine(ctx, path, request, options...)
}

func (vc *VaultClient) MountDisable(ctx context.Context, path string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.System.MountsDisableSecretsEngine(ctx, path, options...)
}

func (vc *VaultClient) ListMounts(ctx context.Context, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.System.MountsListSecretsEngines(ctx, options...)
}
