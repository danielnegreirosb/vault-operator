package cvault

import (
	"context"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
)

type UserPassOperator struct {
	client VaultClientI
}

func NewUserPassOperator(client VaultClientI) *UserPassOperator {
	return &UserPassOperator{client: client}
}

func (uo *UserPassOperator) CreateUser(mountPath string, username string, password string, policies []string, token string) error {
	ok, err := uo.IsUserCreated(uo.client, mountPath, username, token)
	if err != nil {
		return err
	}

	if ok {
		return nil
	}

	_, err = uo.client.CreateUserPass(
		context.Background(),
		username,
		schema.UserpassWriteUserRequest{
			Password: password,
			Policies: policies,
		},
		vault.WithToken(token),
		vault.WithMountPath(mountPath),
	)
	return err
}

func (uo *UserPassOperator) DeleteUserPass(mountPath string, username string, token string) error {
	_, err := uo.client.DeleteUserPass(
		context.Background(),
		username,
		vault.WithToken(token),
		vault.WithMountPath(mountPath),
	)
	return err
}

func (uo *UserPassOperator) IsUserCreated(client VaultClientI, mountPath string, user string, token string) (bool, error) {
	resp, err := client.ListUserPass(context.Background(), vault.WithToken(token), vault.WithMountPath(mountPath))
	if err != nil {
		// If the error is due to the mount path not existing, we consider that the user is not created
		vaultErr, ok := err.(*vault.ResponseError)
		if !ok {
			return false, err
		}

		if vaultErr.StatusCode == 404 && len(vaultErr.Errors) == 0 {
			return false, nil
		}

		return true, err
	}

	for _, key := range resp.Data.Keys {
		if key == user {
			return true, nil
		}
	}

	return false, nil
}

func (vc *VaultClient) CreateUserPass(ctx context.Context, username string, request schema.UserpassWriteUserRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.Auth.UserpassWriteUser(ctx, username, request, options...)
}

func (vc *VaultClient) ListUserPass(ctx context.Context, options ...vault.RequestOption) (*vault.Response[schema.StandardListResponse], error) {
	return vc.Auth.UserpassListUsers(ctx, options...)
}

func (vc *VaultClient) DeleteUserPass(ctx context.Context, username string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.Auth.UserpassDeleteUser(ctx, username, options...)
}
