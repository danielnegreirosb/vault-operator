/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/danielnegreiros/vault-operator/api/v1alpha1"
	cvault "github.com/danielnegreiros/vault-operator/internal/vault/client"
)

const (
	vaultFinalizer     = "vault.finalizers.ops.community.dev"
	defaultRequeueTime = 5 * time.Minute
	errorRequeueTime   = 1 * time.Minute
)

type Phase string

const (
	// Data
	PhaseDataNotValidated Phase = "DataNotValidated"

	// Initialization
	PhaseCheckInitErr Phase = "InitializationUnknown"
	PhaseNotInit      Phase = "NotInitialized"

	// Seal
	PhaseCheckSealErr Phase = "SealStatusUnknown"
	PhaseUnsealErr    Phase = "UnsealError"

	// Secret
	PhaseSaveSecretErr Phase = "SaveSecretFailed"
	PhaseReadSecretErr Phase = "ReadSecretFailed"

	// Connectivity
	PhaseNotReachable Phase = "NotReachable"
	PhaseReachable    Phase = "Reachable"
	PhaseUnsealed     Phase = "Unsealed"
)

type VaultOperatorClient struct {
	Client    cvault.VaultClientI
	Token     string
	Name      string
	Namespace string
	Endpoint  string
}

// VaultServerReconciler reconciles a VaultServer object
type VaultServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=vaultservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=vaultservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=vaultservers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *VaultServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Starting Vault Server Reconciliation")

	obj := &v1alpha1.VaultServer{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, obj)
	}

	// Ensure finalizer (this modifies spec, not status)
	if !controllerutil.ContainsFinalizer(obj, vaultFinalizer) {
		patch := client.MergeFrom(obj.DeepCopy())
		controllerutil.AddFinalizer(obj, vaultFinalizer)
		if err := r.Patch(ctx, obj, patch); err != nil {
			return ctrl.Result{}, err
		}
		// Return immediately after adding finalizer to avoid conflicts
		return ctrl.Result{Requeue: true}, nil
	}

	// Validate Spec
	if err := r.validateSpec(obj); err != nil {
		return r.updateStatus(ctx, obj, PhaseDataNotValidated, err.Error(), errorRequeueTime)
	}

	// Build vault client
	endpoint := buildURL(obj.Spec.Server)
	vaultClient, err := cvault.GetClient(endpoint, cvault.WithTimeout(5))
	if err != nil {
		return r.updateStatus(ctx, obj, PhaseDataNotValidated,
			fmt.Sprintf("failed to build vault client: %v", err), errorRequeueTime)
	}

	vo := cvault.GetVaultOperator(vaultClient, nil)

	// Check connectivity
	if err := vo.Ping(ctx); err != nil {
		return r.updateStatus(ctx, obj, PhaseNotReachable,
			fmt.Sprintf("vault not reachable: %v", err), errorRequeueTime)
	}

	// Handle initialization
	if obj.Spec.Init != nil && *obj.Spec.Init {
		if err := r.handleInitialization(ctx, obj, vo, req); err != nil {
			return r.handleError(ctx, obj, err)
		}
	}

	// Handle auto-unlock
	if obj.Spec.AutoUnlock != nil && *obj.Spec.AutoUnlock {
		if err := r.handleAutoUnlock(ctx, obj, vo); err != nil {
			return r.handleError(ctx, obj, err)
		}
	}

	return r.updateStatus(ctx, obj, PhaseUnsealed, "Vault is operational", defaultRequeueTime)
}

// SetupWithManager sets up the controller with the Manager.
func (r *VaultServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.VaultServer{}).
		Owns(&corev1.Secret{}).
		Named("vaultserver").
		Complete(r)
}

func (r *VaultServerReconciler) handleDeletion(ctx context.Context, obj *v1alpha1.VaultServer) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(obj, vaultFinalizer) {
		logger.Info("Cleaning up VaultServer resources")

		// Perform cleanup (e.g., revoke tokens, cleanup external resources)
		if err := r.cleanup(ctx, obj); err != nil {
			logger.Error(err, "Failed to cleanup resources")
			return ctrl.Result{RequeueAfter: errorRequeueTime}, err
		}

		patch := client.MergeFrom(obj.DeepCopy())
		controllerutil.RemoveFinalizer(obj, vaultFinalizer)
		if err := r.Patch(ctx, obj, patch); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *VaultServerReconciler) cleanup(ctx context.Context, obj *v1alpha1.VaultServer) error {
	// Add cleanup logic here
	// For example: revoke tokens, delete external resources, etc.
	return nil
}

func (r *VaultServerReconciler) validateSpec(obj *v1alpha1.VaultServer) error {
	if obj.Spec.Server.ServiceName == "" {
		return fmt.Errorf("server.serviceName cannot be empty")
	}
	if obj.Spec.Server.Port < 1 || obj.Spec.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	return nil
}

func (r *VaultServerReconciler) handleInitialization(ctx context.Context, obj *v1alpha1.VaultServer, vo interface {
	IsInitialized(context.Context) (bool, error)
	InitVault(context.Context) (map[string]any, error)
}, req ctrl.Request) error {
	logger := log.FromContext(ctx)

	isInit, err := vo.IsInitialized(ctx)
	if err != nil {
		return &vaultError{phase: PhaseCheckInitErr, message: "failed to check initialization status", err: err}
	}

	if !isInit {
		logger.Info("Initializing Vault")
		initData, err := vo.InitVault(ctx)
		if err != nil {
			return &vaultError{phase: PhaseNotInit, message: "failed to initialize vault", err: err}
		}

		if err := r.createOrUpdateSecret(ctx, obj, initData); err != nil {
			return &vaultError{phase: PhaseSaveSecretErr, message: "failed to save init data", err: err}
		}
		logger.Info("Vault initialized successfully")
	}

	return nil
}

func (r *VaultServerReconciler) handleAutoUnlock(ctx context.Context, obj *v1alpha1.VaultServer, vo interface {
	IsSealed(context.Context) (bool, error)
	Unseal(context.Context, []interface{}) error
}) error {
	logger := log.FromContext(ctx)

	isSealed, err := vo.IsSealed(ctx)
	if err != nil {
		return &vaultError{phase: PhaseCheckSealErr, message: "failed to check seal status", err: err}
	}

	if isSealed {
		logger.Info("Vault is sealed, attempting to unseal")
		keys, err := r.getUnsealKeys(ctx, obj)
		if err != nil {
			return &vaultError{phase: PhaseReadSecretErr, message: "failed to retrieve unseal keys", err: err}
		}

		if err := vo.Unseal(ctx, keys); err != nil {
			return &vaultError{phase: PhaseUnsealErr, message: "failed to unseal vault", err: err}
		}
		logger.Info("Vault unsealed successfully")
	}

	return nil
}

func buildURL(config v1alpha1.VaultServerConfig) string {
	var b strings.Builder

	// Use HTTPS if TLS is configured, otherwise HTTP
	scheme := "http"
	// if spec.Server.TLS != nil && *spec.Server.TLS {
	// 	scheme = "https"
	// }
	b.WriteString(scheme)
	b.WriteString("://")
	b.WriteString(config.ServiceName)

	if ns := config.Namespace; ns != "" {
		b.WriteString(".")
		b.WriteString(ns)
		b.WriteString(".svc.cluster.local")
	}

	if port := config.Port; port > 0 {
		fmt.Fprintf(&b, ":%d", port)
	}

	return b.String()
}

func (r *VaultServerReconciler) getUnsealKeys(ctx context.Context, obj *v1alpha1.VaultServer) ([]interface{}, error) {
	secret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      obj.Name + "-secret",
		Namespace: obj.Namespace,
	}, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Extract and sort keys
	type keyPair struct {
		index int
		value string
	}
	var keys []keyPair

	for k, v := range secret.Data {
		if k == "root_token" {
			continue
		}
		index, err := strconv.Atoi(k)
		if err != nil {
			continue // Skip non-numeric keys
		}
		keys = append(keys, keyPair{index: index, value: string(v)})
	}

	// Sort by index
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].index < keys[j].index
	})

	// Convert to []interface{}
	result := make([]interface{}, len(keys))
	for i, kp := range keys {
		result[i] = kp.value
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no unseal keys found in secret")
	}

	return result, nil
}

func (r *VaultServerReconciler) createOrUpdateSecret(ctx context.Context, obj *v1alpha1.VaultServer, data map[string]interface{}) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name + "-secret",
			Namespace: obj.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		// Extract root token
		rootToken, ok := data["root_token"].(string)
		if !ok {
			return fmt.Errorf("root_token not found or invalid type in init data")
		}

		// Extract keys
		keys, ok := data["keys"].([]interface{})
		if !ok {
			return fmt.Errorf("keys not found or invalid type in init data")
		}

		// Build string data
		stringData := map[string]string{
			"root_token": rootToken,
		}
		for i, k := range keys {
			keyStr, ok := k.(string)
			if !ok {
				return fmt.Errorf("key at index %d is not a string", i)
			}
			stringData[strconv.Itoa(i+1)] = keyStr
		}

		secret.Type = corev1.SecretTypeOpaque
		secret.StringData = stringData

		// Set owner reference
		return controllerutil.SetControllerReference(obj, secret, r.Scheme)
	})

	return err
}

func (r *VaultServerReconciler) updateStatus(ctx context.Context, obj *v1alpha1.VaultServer, phase Phase, message string, requeueAfter time.Duration) (ctrl.Result, error) {
	// Use a retry loop for status updates to handle conflicts
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Fetch the latest version
		latest := &v1alpha1.VaultServer{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(obj), latest); err != nil {
			return err
		}

		// Update status on the latest version
		latest.Status.Phase = string(phase)
		latest.Status.Message = message
		latest.Status.LastUpdateTime = &metav1.Time{Time: time.Now()}

		return r.Status().Update(ctx, latest)
	})

	if err != nil {
		log.Log.Error(err, "Failed to update VaultServer status after retries")
		return ctrl.Result{RequeueAfter: errorRequeueTime}, err
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

func (r *VaultServerReconciler) handleError(ctx context.Context, obj *v1alpha1.VaultServer, err error) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if vErr, ok := err.(*vaultError); ok {
		logger.Error(vErr.err, vErr.message, "phase", vErr.phase)
		return r.updateStatus(ctx, obj, vErr.phase, vErr.message, errorRequeueTime)
	}

	logger.Error(err, "Unexpected error during reconciliation")
	return r.updateStatus(ctx, obj, PhaseDataNotValidated, err.Error(), errorRequeueTime)
}

func getVaultOpClient(ctx context.Context, name string, namespace string, reqNamespace string, client client.Client) (*VaultOperatorClient, error) {

	vaultOpSearch := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	if vaultOpSearch.Namespace == "" {
		vaultOpSearch.Namespace = reqNamespace
	}

	vaultOpInstance := &v1alpha1.VaultServer{}
	if err := client.Get(ctx, vaultOpSearch, vaultOpInstance); err != nil {
		return nil, fmt.Errorf("failed to get vault operator instance: %v", err)
	}

	url := buildURL(vaultOpInstance.Spec.Server)

	// Build vault client
	// endpoint := buildURL(vaultObj.Spec)
	vaultClient, err := cvault.GetClient(url, cvault.WithTimeout(5))
	if err != nil {
		return nil, fmt.Errorf("failed to connect with vault: %v", err)
	}

	vaultOpSearch.Name += "-secret"
	vaultToken := &corev1.Secret{}
	if err := client.Get(ctx, vaultOpSearch, vaultToken); err != nil {
		return nil, err
	}

	return &VaultOperatorClient{
		Name:      vaultOpInstance.Name,
		Namespace: vaultOpInstance.Namespace,
		Client:    vaultClient,
		Token:     string(vaultToken.Data["root_token"]),
		Endpoint:  url,
	}, nil
}

// vaultError is a custom error type that includes phase information
type vaultError struct {
	phase   Phase
	message string
	err     error
}

func (e *vaultError) Error() string {
	return fmt.Sprintf("%s: %v", e.message, e.err)
}

func (e *vaultError) Unwrap() error {
	return e.err
}
