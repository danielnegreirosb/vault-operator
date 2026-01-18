package cvault

import (
	"context"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"

	vapi "github.com/hashicorp/vault/api"
)

type VaultClientI interface {
	// System
	ReadInitializationStatus(ctx context.Context, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)
	Initialize(ctx context.Context, request schema.InitializeRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)
	SealStatus(ctx context.Context, options ...vault.RequestOption) (*vault.Response[schema.SealStatusResponse], error)
	Unseal(ctx context.Context, request schema.UnsealRequest, options ...vault.RequestOption) (*vault.Response[schema.UnsealResponse], error)
	ReadHealthStatus(ctx context.Context, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)

	// Secrets
	MountsEnableSecretsEngine(ctx context.Context, path string, request schema.MountsEnableSecretsEngineRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)
	KvV2Read(ctx context.Context, path string, options ...vault.RequestOption) (*vault.Response[schema.KvV2ReadResponse], error)
	KvV2Write(ctx context.Context, path string, request schema.KvV2WriteRequest, options ...vault.RequestOption) (*vault.Response[schema.KvV2WriteResponse], error)
	KvV2Delete(ctx context.Context, path string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)

	// Policies
	PoliciesWriteAclPolicy(ctx context.Context, name string, request schema.PoliciesWriteAclPolicyRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)
	PoliciesDeleteAclPolicy(ctx context.Context, name string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)

	// Auth Methods
	AuthEnableAuthMethod(ctx context.Context, path string, request schema.AuthEnableMethodRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)
	AuthDisableAuthMethod(ctx context.Context, path string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)
	ListAuthMethods(ctx context.Context, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)

	// Secret Engines
	ListMounts(ctx context.Context, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)
	MountDisable(ctx context.Context, path string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)
	MountEnable(ctx context.Context, path string, request schema.MountsEnableSecretsEngineRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)

	// Userpass Auth Method
	CreateUserPass(ctx context.Context, username string, request schema.UserpassWriteUserRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)
	ListUserPass(ctx context.Context, options ...vault.RequestOption) (*vault.Response[schema.StandardListResponse], error)
	DeleteUserPass(ctx context.Context, username string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)

	// AppRole Auth Method
	CreateAppRoleService(ctx context.Context, roleName string, secretIDTTL string, policies []string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)
	WriteAppRoleWithContext(ctx context.Context, path string, roleName string, data map[string]interface{}, ep string, token string) (*vapi.Secret, error)
	GetAppRole(ctx context.Context, roleName string, options ...vault.RequestOption) (*vault.Response[schema.AppRoleReadRoleIdResponse], error)
	DeleteAppRole(ctx context.Context, roleName string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)
}

type VaultClient struct {
	*vault.Client
}

var _ VaultClientI = (*VaultClient)(nil)

func (vc *VaultClient) ReadInitializationStatus(ctx context.Context, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.System.ReadInitializationStatus(ctx)
}

func (vc *VaultClient) Initialize(ctx context.Context, request schema.InitializeRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.System.Initialize(ctx, request, options...)
}

func (vc *VaultClient) SealStatus(ctx context.Context, options ...vault.RequestOption) (*vault.Response[schema.SealStatusResponse], error) {
	return vc.System.SealStatus(ctx, options...)
}

func (vc *VaultClient) Unseal(ctx context.Context, request schema.UnsealRequest, options ...vault.RequestOption) (*vault.Response[schema.UnsealResponse], error) {
	return vc.System.Unseal(ctx, request, options...)
}

func (vc *VaultClient) MountsEnableSecretsEngine(ctx context.Context, path string, request schema.MountsEnableSecretsEngineRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.System.MountsEnableSecretsEngine(ctx, path, request)
}

func (vc *VaultClient) ReadHealthStatus(ctx context.Context, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.System.ReadHealthStatus(ctx, options...)
}

func (vc *VaultClient) KvV2Read(ctx context.Context, path string, options ...vault.RequestOption) (*vault.Response[schema.KvV2ReadResponse], error) {
	return vc.Secrets.KvV2Read(ctx, path, options...)
}

func (vc *VaultClient) KvV2Write(ctx context.Context, path string, request schema.KvV2WriteRequest, options ...vault.RequestOption) (*vault.Response[schema.KvV2WriteResponse], error) {
	return vc.Secrets.KvV2Write(ctx, path, request, options...)
}

func (vc *VaultClient) KvV2Delete(ctx context.Context, path string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.Secrets.KvV2Delete(ctx, path, options...)
}

type VaultOption func() vault.ClientOption

func GetClient(url string, options ...VaultOption) (VaultClientI, error) {

	clientOptions := []vault.ClientOption{vault.WithAddress(url)}
	for _, vo := range options {
		clientOptions = append(clientOptions, vo())
	}

	client, err := vault.New(
		clientOptions...,
	)

	if err != nil {
		return nil, err
	}

	return &VaultClient{Client: client}, nil
}

type VaultOperator struct {
	secretOperator *SecretOperator
	client         VaultClientI
}

func GetVaultOperator(vclient VaultClientI, secretOp *SecretOperator) *VaultOperator {

	vc := &VaultOperator{client: vclient, secretOperator: secretOp}
	return vc
}

func WithTimeout(timeout int) VaultOption {
	return func() vault.ClientOption {
		return vault.WithRequestTimeout(time.Duration(timeout) * time.Second)
	}
}
