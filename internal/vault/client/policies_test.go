package cvault

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/stretchr/testify/assert"
)

func TestPolicyCreation(t *testing.T) {

	ctx := context.Background()

	policies := []string{
		`path "unlocker/data/*" { capabilities = [ "read", "list" ]}`,
		`path "unlocker/metadata/*" { capabilities = [ "read", "list" ]}`,
	}

	client := &MockVaultClient{}
	policyOp := NewPoliciesOperator(client)

	err := policyOp.CreateOrUpdateAclPolicy(ctx, "my-policy", policies, "token")
	assert.NoError(t, err)
	assert.Equal(t, client.policyCount, 2)

}

func TestPolicyDeletion(t *testing.T) {

	ctx := context.Background()

	client := &MockVaultClient{}
	policyOp := NewPoliciesOperator(client)

	err := policyOp.DeleteAclPolicy(ctx, "my-policy", "token")
	assert.NoError(t, err)

}

func (vc *MockVaultClient) PoliciesWriteAclPolicy(ctx context.Context, name string, request schema.PoliciesWriteAclPolicyRequest, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	vc.policyCount = len(strings.Split(request.Policy, "\r\n"))
	return nil, nil
}

func (vc *MockVaultClient) PoliciesDeleteAclPolicy(ctx context.Context, name string, options ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	return nil, nil
}
