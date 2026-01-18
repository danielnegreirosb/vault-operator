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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/danielnegreiros/vault-operator/api/v1alpha1"
	vaultv1alpha1 "github.com/danielnegreiros/vault-operator/api/v1alpha1"
	cvault "github.com/danielnegreiros/vault-operator/internal/vault/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	appRoleFinalizer = "approle.finalizers.ops.community.dev"
)

// AppRoleReconciler reconciles a AppRole object
type AppRoleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=approles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=approles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=approles/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppRole object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *AppRoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := logf.FromContext(ctx)

	logger.Info("Reconciling AppRole")

	// Fetch the AppRole instance
	appRole := &v1alpha1.AppRole{}
	if err := r.Get(ctx, req.NamespacedName, appRole); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("AppRole resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get AppRole")
		return ctrl.Result{}, err
	}

	vaultOpInstance, err := getVaultOpClient(ctx, appRole.Spec.VaultServer.Name, appRole.Spec.VaultServer.Namespace,
		req.Namespace, r.Client)
	if err != nil {
		return r.updateStatus(ctx, appRole, false,
			fmt.Sprintf("Failed to get vault operator client: %v", err), errorRequeueTime)
	}

	appOp := cvault.NewAppRoleOperator(vaultOpInstance.Client, vaultOpInstance.Endpoint)

	if !appRole.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, appRole, appOp, vaultOpInstance.Token)
	
	}

	if !controllerutil.ContainsFinalizer(appRole, appRoleFinalizer) {
		patch := client.MergeFrom(appRole.DeepCopy())
		controllerutil.AddFinalizer(appRole, appRoleFinalizer)
		if err := r.Patch(ctx, appRole, patch); err != nil {
			return r.updateStatus(ctx, appRole, false,
				fmt.Sprintf("Failed to add finalizer to AppRole: %v", err), errorRequeueTime)
		}
		// Return immediately after adding finalizer to avoid conflicts
		return ctrl.Result{RequeueAfter: 0}, nil
	}

	err = appOp.CreateorUpdateAppRole(ctx, appRole.Spec.MountPath, appRole.Spec.Name,
		appRole.Spec.SecretIDTTL, appRole.Spec.Policies, vaultOpInstance.Token)
	if err != nil {
		return r.updateStatus(ctx, appRole, false,
			fmt.Sprintf("Failed to create or update AppRole in Vault: %v", err), errorRequeueTime)
	}

	if appRole.Spec.Export != nil && appRole.Spec.Export.Namespace != "" {
		roleId, err := appOp.GetRoleId(ctx, appRole.Spec.Name, appRole.Spec.MountPath, vaultOpInstance.Token)
		if err != nil {
			return r.updateStatus(ctx, appRole, false, fmt.Sprintf("Failed to get AppRole RoleId: %v", err), errorRequeueTime)
		}

		secretId, err := appOp.GenerateAppRoleSecretID(ctx, appRole.Spec.MountPath, appRole.Spec.Name, vaultOpInstance.Token)
		if err != nil {
			return r.updateStatus(ctx, appRole, false, fmt.Sprintf("Failed to generate AppRole SecretId: %v", err), errorRequeueTime)
		}

		secretName := fmt.Sprintf("approle-%s-secret", appRole.Name)
		err = r.exportAppRoleSecret(ctx, secretName, roleId, secretId, appRole.Spec.Export.Namespace)
		if err != nil {
			return r.updateStatus(ctx, appRole, false,
				fmt.Sprintf("Failed to export AppRole secret: %v", err), errorRequeueTime)
		}
	}

	return r.updateStatus(ctx, appRole, true,
		"AppRole successfully synchronized", 0)
}

func (r *AppRoleReconciler) handleDeletion(ctx context.Context, appRole *v1alpha1.AppRole, appOp *cvault.AppRoleOperator, token string) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Info("Handling deletion of AppRole")

	if controllerutil.ContainsFinalizer(appRole, appRoleFinalizer) {
		err := appOp.DeleteAppRole(ctx, appRole.Spec.MountPath, appRole.Spec.Name, token)
		if err != nil {
			return ctrl.Result{RequeueAfter: errorRequeueTime}, fmt.Errorf("failed to delete AppRole from Vault: %v", err)
		}

		patch := client.MergeFrom(appRole.DeepCopy())
		controllerutil.RemoveFinalizer(appRole, appRoleFinalizer)
		if err := r.Patch(ctx, appRole, patch); err != nil {
			// no requeue needed here as we're deleting
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *AppRoleReconciler) exportAppRoleSecret(ctx context.Context, secretName, roleId, secretId, namespace string) error {

	secretObj := constructAppRoleSecretObject(secretName, namespace, roleId, secretId)

	err := createOrUpdateK8sSecret(ctx, r.Client, secretObj)
	if err != nil {
		return fmt.Errorf("failed to create or update exported AppRole secret: %v", err)
	}

	return nil
}

func constructAppRoleSecretObject(name, namespace string, roleId, secretId string) *corev1.Secret {
	data := make(map[string][]byte)
	data["role_id"] = []byte(roleId)
	data["secret_id"] = []byte(secretId)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}

func createOrUpdateK8sSecret(ctx context.Context, k8sClient client.Client, secret *corev1.Secret) error {
	existingSecret := &corev1.Secret{}
	err := k8sClient.Get(ctx, client.ObjectKey{Name: secret.Name, Namespace: secret.Namespace}, existingSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create the secret
			return k8sClient.Create(ctx, secret)
		}
		return err
	}

	// Update the existing secret
	existingSecret.Data = secret.Data
	return k8sClient.Update(ctx, existingSecret)
}

func (r *AppRoleReconciler) updateStatus(ctx context.Context, appRole *v1alpha1.AppRole,
	synced bool, message string, requeueAfter time.Duration) (ctrl.Result, error) {

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest := &v1alpha1.AppRole{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(appRole), latest); err != nil {
			return err
		}

		latest.Status.Synchronized = strconv.FormatBool(synced)
		latest.Status.Message = message
		latest.Status.LastUpdateTime = &metav1.Time{Time: time.Now()}

		return r.Status().Update(ctx, latest)
	})

	if err != nil {
		return ctrl.Result{RequeueAfter: errorRequeueTime}, err
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *AppRoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vaultv1alpha1.AppRole{}).
		Named("approle").
		Complete(r)
}
