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
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/danielnegreiros/vault-operator/api/v1alpha1"
	vaultv1alpha1 "github.com/danielnegreiros/vault-operator/api/v1alpha1"
	cvault "github.com/danielnegreiros/vault-operator/internal/vault/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	authMethodFinalizer = "authmethod.finalizers.ops.community.dev"
)

// AuthMethodReconciler reconciles a AuthMethod object
type AuthMethodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=authmethods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=authmethods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=authmethods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AuthMethod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *AuthMethodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Info("Starting Auth Method Reconciliation")

	obj := &v1alpha1.AuthMethod{}
	err := r.Get(ctx, req.NamespacedName, obj)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	vaultOpInstance, err := getVaultOpClient(ctx, obj.Spec.VaultServer.Name, obj.Spec.VaultServer.Namespace, req.Namespace, r.Client)
	if err != nil {
		return r.updateStatus(ctx, obj, false,
			fmt.Sprintf("Failed to get vault operator client: %v", err), errorRequeueTime)
	}

	vaultAuthOp := cvault.NewAuthOperator(vaultOpInstance.Client)
	if !obj.GetDeletionTimestamp().IsZero() {
		return r.handleDeletion(ctx, obj, vaultAuthOp, vaultOpInstance.Token)
	}

	if !controllerutil.ContainsFinalizer(obj, authMethodFinalizer) {
		patch := client.MergeFrom(obj.DeepCopy())
		controllerutil.AddFinalizer(obj, authMethodFinalizer)
		if err := r.Patch(ctx, obj, patch); err != nil {
			return r.updateStatus(ctx, obj, false,
				fmt.Sprintf("Failed to add finalizer: %v", err), errorRequeueTime)
		}

		return ctrl.Result{RequeueAfter: 0}, nil
	}

	err = vaultAuthOp.EnableAuthMethod(obj.Spec.Path, obj.Spec.Type, vaultOpInstance.Token)
	if err != nil {
		return r.updateStatus(ctx, obj, false,
			fmt.Sprintf("Failed to enable auth method: %v", err), errorRequeueTime)
	}

	return r.updateStatus(ctx, obj, true,
		"Auth method synchronized successfully", defaultRequeueTime)

}

func (r *AuthMethodReconciler) updateStatus(ctx context.Context, obj *vaultv1alpha1.AuthMethod, synchronized bool, message string, requeueAfter time.Duration) (ctrl.Result, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest := &v1alpha1.AuthMethod{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(obj), latest); err != nil {
			return err
		}

		latest.Status.Message = message
		latest.Status.Synchronized = strconv.FormatBool(synchronized)
		latest.Status.LastUpdateTime = &metav1.Time{Time: time.Now()}

		return r.Status().Update(ctx, latest)
	})

	logger := logf.FromContext(ctx)

	if err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{RequeueAfter: errorRequeueTime}, err
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil

}

func (r *AuthMethodReconciler) handleDeletion(ctx context.Context, obj *v1alpha1.AuthMethod, vaultAuthOp *cvault.AuthOperator, token string) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(obj, authMethodFinalizer) {
		// Our finalizer is present, so lets handle any external dependency
		err := vaultAuthOp.DisableAuthMethod(obj.Spec.Path, token)
		if err != nil {
			// requeue on error for proper cleaning
			return ctrl.Result{RequeueAfter: errorRequeueTime}, err
		}

		patch := client.MergeFrom(obj.DeepCopy())
		controllerutil.RemoveFinalizer(obj, authMethodFinalizer)

		if err := r.Patch(ctx, obj, patch); err != nil {
			// no requeue for a deletion patch error
			return ctrl.Result{}, err
		}
	}

	// Stop reconciliation as the item is deleted
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AuthMethodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vaultv1alpha1.AuthMethod{}).
		Named("authmethod").
		Complete(r)
}
