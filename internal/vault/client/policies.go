package cvault

import (
	"context"
	"strings"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PoliciesOperator struct {
	client VaultClientI
}

func NewPoliciesOperator(client VaultClientI) *PoliciesOperator {
	return &PoliciesOperator{client: client}
}

func (so *PoliciesOperator) CreateOrUpdateAclPolicy(ctx context.Context, name string, rules []string, token string) error {
	logger := log.FromContext(ctx)
	logger.Info("Starting policy creation or update", "name", name)

	rulesString := strings.Join(rules, "\r\n")

	_, err := so.client.PoliciesWriteAclPolicy(ctx, name, schema.PoliciesWriteAclPolicyRequest{
		Policy: rulesString,
	}, vault.WithToken(token))
	return err
}

func (so *PoliciesOperator) DeleteAclPolicy(ctx context.Context, name string, token string) error {
	logger := log.FromContext(ctx)
	logger.Info("Starting policy deletion", "name", name)

	_, err := so.client.PoliciesDeleteAclPolicy(ctx, name, vault.WithToken(token))
	return err
}

func (vc *VaultClient) PoliciesWriteAclPolicy(ctx context.Context, name string, request schema.PoliciesWriteAclPolicyRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.System.PoliciesWriteAclPolicy(ctx, name, request, options...)
}

func (vc *VaultClient) PoliciesDeleteAclPolicy(ctx context.Context, name string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return vc.System.PoliciesDeleteAclPolicy(ctx, name, options...)
}
