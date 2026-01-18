package cvault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (v *VaultOperator) IsInitialized(ctx context.Context) (bool, error) {
	resp, err := v.client.ReadInitializationStatus(ctx)
	if err != nil {
		return true, err
	}

	return resp.Data["initialized"].(bool), nil

}

func (v *VaultOperator) InitVault(ctx context.Context) (map[string]interface{}, error) {
	resp, err := v.client.Initialize(ctx, schema.InitializeRequest{
		SecretShares:    3,
		SecretThreshold: 3,
	})

	if err != nil {
		return nil, fmt.Errorf("vault init: [%w]", err)
	}

	return resp.Data, nil
}

func (v *VaultOperator) IsSealed(ctx context.Context) (bool, error) {

	status, err := v.client.SealStatus(ctx)
	if err != nil {
		return true, err
	}

	return status.Data.Sealed, nil

}

func (v *VaultOperator) Unseal(ctx context.Context, keys []interface{}) error {
	for _, k := range keys {
		_, err := v.client.Unseal(ctx, schema.UnsealRequest{
			Key: k.(string),
		})

		if err != nil {
			return fmt.Errorf("unseal: [%w]", err)
		}
	}

	return nil
}

func (v *VaultOperator) mountKvEnginePath(ctx context.Context, path string, kvType string, token string) error {
	logger := log.FromContext(ctx)

	_, err := v.client.MountsEnableSecretsEngine(ctx, path, schema.MountsEnableSecretsEngineRequest{
		Type: kvType,
	}, vault.WithToken(token))
	if err != nil {
		return fmt.Errorf("enable kv [%w]", err)
	}

	logger.Info("enable kv operation completed", "type", kvType, "mountPath", path)
	return nil
}

func (v *VaultOperator) Ping(ctx context.Context) error {
	_, err := v.client.ReadHealthStatus(ctx)
	return err
}
