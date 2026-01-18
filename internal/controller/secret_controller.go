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
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/danielnegreiros/vault-operator/api/v1alpha1"
	vaultv1alpha1 "github.com/danielnegreiros/vault-operator/api/v1alpha1"
	cvault "github.com/danielnegreiros/vault-operator/internal/vault/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	secretFinalizer = "secret.finalizers.ops.community.dev"
)

// SecretReconciler reconciles a Secret object
type SecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=secrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=secrets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Secret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	logger := logf.FromContext(ctx)
	logger.Info("Reconciling Secret")

	secret := &vaultv1alpha1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	vaultOpInstance, err := getVaultOpClient(ctx, secret.Spec.VaultServer.Name, secret.Spec.VaultServer.Namespace, req.Namespace, r.Client)
	if err != nil {
		return r.updateSecretStatus(ctx, secret, false,
			fmt.Sprintf("Failed to get vault operator client: %v", err), errorRequeueTime)
	}

	so := cvault.NewSecretOperator(vaultOpInstance.Client)

	if !secret.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleSecretDeletion(ctx, secret, so, vaultOpInstance.Token)
	}

	// Ensure finalizer (this modifies spec, not status)
	if !controllerutil.ContainsFinalizer(secret, secretFinalizer) {
		patch := client.MergeFrom(secret.DeepCopy())
		controllerutil.AddFinalizer(secret, secretFinalizer)
		if err := r.Patch(ctx, secret, patch); err != nil {
			return ctrl.Result{}, err
		}
		// Return immediately after adding finalizer to avoid conflicts
		return ctrl.Result{RequeueAfter: 0}, nil
	}

	err = so.CreateOrUpdateKvV2Secret(ctx, secret.Spec.MountPath, secret.Spec.Path, secret.Spec.Name, vaultOpInstance.Token, secret.Spec.Data)
	if err != nil {
		return r.updateSecretStatus(ctx, secret, false,
			fmt.Sprintf("Not possible to create/update secret at path: %s", secret.Spec.Path), errorRequeueTime)
	}

	return r.updateSecretStatus(ctx, secret, true, "Sync", defaultRequeueTime)
}

func (r *SecretReconciler) updateSecretStatus(ctx context.Context, obj *v1alpha1.Secret, synced bool, message string, requeueAfter time.Duration) (ctrl.Result, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest := &v1alpha1.Secret{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(obj), latest); err != nil {
			return err
		}

		// Update status on latest version
		latest.Status.Synchronized = strconv.FormatBool(synced)
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

func (r *SecretReconciler) handleSecretDeletion(ctx context.Context, obj *v1alpha1.Secret, so *cvault.SecretOperator, token string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(obj, secretFinalizer) {
		logger.Info("Cleaning up", "secret", obj.Spec.Name)

		if err := so.DeleteKvV2Secret(ctx, obj.Spec.MountPath, obj.Spec.Path, obj.Spec.Name, token); err != nil {
			logger.Error(err, "Error cleaning up", "secret", obj.Spec.Name)
			return ctrl.Result{RequeueAfter: errorRequeueTime}, err
		}

		patch := client.MergeFrom(obj.DeepCopy())
		controllerutil.RemoveFinalizer(obj, secretFinalizer)
		if err := r.Patch(ctx, obj, patch); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vaultv1alpha1.Secret{}).
		Named("secret").
		Complete(r)
}
