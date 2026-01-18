package cvault

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	randv2 "math/rand/v2"
	"net/url"
	"strings"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	errManipulatingSecretPath = errors.New("error manipulating secret path")
	errAddingSecret           = errors.New("error when adding secret")
	errCheckingSecretExists   = errors.New("not possible to check if secret exists")

	errorFormat = "%w mount=%s path=%s secret=%s: %w"
)

type SecretOperator struct {
	client VaultClientI
}

func NewSecretOperator(client VaultClientI) *SecretOperator {
	return &SecretOperator{client: client}
}

func (so *SecretOperator) DeleteKvV2Secret(ctx context.Context, mountPath string, secretPath string, name string, token string) error {
	logger := log.FromContext(ctx)
	logger.Info("Starting secret deletion", "mount", mountPath, "secret_path", secretPath, "name", name)

	secretPathName, err := url.JoinPath(secretPath, name)
	if err != nil {
		return fmt.Errorf(errorFormat, errManipulatingSecretPath, mountPath, secretPath, name, err)
	}

	_, err = so.client.KvV2Delete(ctx, secretPathName, vault.WithMountPath(mountPath), vault.WithToken(token))
	if err != nil {
		return fmt.Errorf(errorFormat, errAddingSecret, mountPath, secretPath, name, err)
	}

	logger.Info("Completed secret deletion", "mount", mountPath, "secret_path", secretPath, "name", name)
	return nil
}

func (so *SecretOperator) CreateOrUpdateKvV2Secret(ctx context.Context, mountPath string, secretPath string, name string, token string, data map[string]string) error {
	logger := log.FromContext(ctx)
	logger.Info("Starting secret creation", "mount", mountPath, "secret_path", secretPath, "name", name)

	secretPathName, err := url.JoinPath(secretPath, name)
	if err != nil {
		return fmt.Errorf("%w: mount=%s path=%s secret=%s: %w", errManipulatingSecretPath, mountPath, secretPath, name, err)
	}

	err = so.checkKVSecretExists(ctx, mountPath, secretPathName, token)
	if err == nil {
		logger.Info("secret already exists, continuing....", "mount", mountPath, "secret", secretPathName)
		return nil
	}

	if strings.Contains(err.Error(), "404") {
		err = so.createOrUpdateKvV2Secret(ctx, secretPathName, mountPath, randomize(data, 32), token)
		if err != nil {
			return fmt.Errorf("%w mount=%s path=%s secret=%s: %w", errAddingSecret, mountPath, secretPath, name, err)
		}
	} else {
		return fmt.Errorf(errorFormat, errCheckingSecretExists, mountPath, secretPath, name, err)
	}

	logger.Info("Completed secret creation", "mount", mountPath, "secret_path", secretPath, "name", name)
	return nil
}

func (so *SecretOperator) checkKVSecretExists(ctx context.Context, mountPath string, path string, token string) error {
	logger := log.FromContext(ctx)
	logger.Info("checking if secret is existent", "mount", mountPath, "path", path)
	_, err := so.client.KvV2Read(ctx, path, vault.WithMountPath(mountPath), vault.WithToken(token))
	return err
}

func (so *SecretOperator) createOrUpdateKvV2Secret(ctx context.Context, secretPath string, mountPath string, data map[string]interface{}, token string) error {
	_, err := so.client.KvV2Write(ctx, secretPath, schema.KvV2WriteRequest{
		Data: data,
	}, vault.WithToken(token),
		vault.WithMountPath(mountPath))
	return err
}

func randomize(m map[string]string, size int) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		if v == "{{random}}" {
			result[k] = generateRandomString(size)
			continue
		}
		result[k] = v
	}
	return result
}

// Helper function to generate a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		for i := range result {
			result[i] = charset[randv2.IntN(len(charset))]
		}
		return string(result)
	}

	for i, b := range randomBytes {
		result[i] = charset[int(b)%len(charset)]
	}
	return string(result)
}
