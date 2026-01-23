package cvault

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	vapi "github.com/hashicorp/vault/api"
)

type AppRoleOperator struct {
	endPoint string
	client   VaultClientI
}

func NewAppRoleOperator(client VaultClientI, ep string) *AppRoleOperator {
	return &AppRoleOperator{client: client, endPoint: ep}
}

func (uo *AppRoleOperator) CreateorUpdateAppRole(ctx context.Context, mountPath string, roleName string, secretIDTTL int, policies []string, token string) error {
	_, err := uo.client.CreateAppRoleService(ctx, roleName, strconv.Itoa(secretIDTTL), policies, vault.WithMountPath(mountPath), vault.WithToken(token))
	if err != nil {
		return err
	}
	return nil
}

func (uo *AppRoleOperator) GenerateAppRoleSecretID(ctx context.Context, path string, roleName string, token string) (string, error) {

	secret, err := uo.client.WriteAppRoleWithContext(ctx, path, roleName, nil, uo.endPoint, token)
	if err != nil {
		return "", err
	}

	// Extract SecretID
	if secret != nil && secret.Data != nil {
		return secret.Data["secret_id"].(string), nil
	} else {
		return "", fmt.Errorf("no secret data found")
	}
}

func (ap *AppRoleOperator) GetRoleId(ctx context.Context, roleName string, mountPath string, token string) (string, error) {
	res, err := ap.client.GetAppRole(ctx, roleName, vault.WithMountPath(mountPath), vault.WithToken(token))
	if err != nil {
		return "", err
	}

	return res.Data.RoleId, nil
}

func (ao *AppRoleOperator) IsAppRoleCreated(ctx context.Context, mountPath string, roleName string, token string) (bool, error) {
	_, err := ao.client.GetAppRole(ctx, roleName, vault.WithMountPath(mountPath), vault.WithToken(token))
	if err != nil {
		if vault.IsErrorStatus(err, 404) {
			return false, nil
		}
		return false, err
	}
	return true, nil

}

func (ao *AppRoleOperator) DeleteAppRole(ctx context.Context, mountPath string, roleName string, token string) error {
	_, err := ao.client.DeleteAppRole(ctx, roleName, vault.WithMountPath(mountPath), vault.WithToken(token))
	if err != nil {
		return err
	}
	return nil
}

func (vc *VaultClient) CreateAppRoleService(ctx context.Context, roleName string, secretIDTTL string, policies []string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.Auth.AppRoleWriteRole(ctx, roleName, schema.AppRoleWriteRoleRequest{
		SecretIdTtl: secretIDTTL,
		Policies:    policies,
	}, options...)
}

func (vc *VaultClient) WriteAppRoleWithContext(ctx context.Context, path string, roleName string, data map[string]interface{}, ep string, token string) (*vapi.Secret, error) {
	config := vapi.DefaultConfig()
	config.Address = ep

	client, err := vapi.NewClient(config)
	if err != nil {
		return nil, err
	}

	client.SetToken(token)
	request := fmt.Sprintf("auth/%s/role/%s/secret-id", path, roleName)

	return client.Logical().WriteWithContext(ctx, request, nil)
}

func (vc *VaultClient) GetAppRole(ctx context.Context, roleName string, options ...vault.RequestOption) (*vault.Response[schema.AppRoleReadRoleIdResponse], error) {
	return vc.Auth.AppRoleReadRoleId(ctx, roleName, options...)
}

func (vc *VaultClient) DeleteAppRole(ctx context.Context, roleName string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.Auth.AppRoleDeleteRole(ctx, roleName, options...)
}
