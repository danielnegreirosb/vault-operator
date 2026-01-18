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
	userPassFinalizer = "userpass.finalizers.ops.community.dev"
)

// UserPassReconciler reconciles a UserPass object
type UserPassReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=userpasses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=userpasses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=userpasses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the UserPass object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *UserPassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Info("Reconciling UserPass")

	obj := &v1alpha1.UserPass{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		logger.Error(err, "unable to fetch UserPass")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	vaultOpInstance, err := getVaultOpClient(ctx, obj.Spec.VaultServer.Name, obj.Spec.VaultServer.Namespace, req.Namespace, r.Client)
	if err != nil {
		return r.updateStatus(ctx, obj, false, fmt.Sprintf("Failed to get vault client: %v", err), errorRequeueTime)
	}

	vaultUserOp := cvault.NewUserPassOperator(vaultOpInstance.Client)
	if !obj.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, obj, *vaultUserOp, vaultOpInstance.Token)
	}

	// Ensure finalizer is present
	if !controllerutil.ContainsFinalizer(obj, userPassFinalizer) {
		patch := client.MergeFrom(obj.DeepCopy())
		controllerutil.AddFinalizer(obj, userPassFinalizer)
		if err := r.Patch(ctx, obj, patch); err != nil {
			return r.updateStatus(ctx, obj, false,
				fmt.Sprintf("Failed to add finalizer: %v", err), errorRequeueTime)
		}
		// avoid updates conflict
		return ctrl.Result{RequeueAfter: 0}, nil
	}

	err = vaultUserOp.CreateUser(obj.Spec.MountPath, obj.Spec.Name, obj.Spec.Password, obj.Spec.Policies, vaultOpInstance.Token)
	if err != nil {
		return r.updateStatus(ctx, obj, false,
			fmt.Sprintf("Failed to create/update user: %v", err), errorRequeueTime)
	}

	return r.updateStatus(ctx, obj, true,
		"User synchronized successfully", defaultRequeueTime)

}

// SetupWithManager sets up the controller with the Manager.
func (r *UserPassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vaultv1alpha1.UserPass{}).
		Named("userpass").
		Complete(r)
}

func (r *UserPassReconciler) updateStatus(ctx context.Context, obj *v1alpha1.UserPass, sync bool, message string, requeueAfter time.Duration) (ctrl.Result, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest := &v1alpha1.UserPass{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(obj), latest); err != nil {
			return err
		}

		latest.Status.Message = message
		latest.Status.Synchronized = strconv.FormatBool(sync)
		latest.Status.LastUpdateTime = &metav1.Time{Time: time.Now()}
		return r.Status().Update(ctx, latest)
	})

	logger := logf.FromContext(ctx)

	if err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

func (r *UserPassReconciler) handleDeletion(ctx context.Context, obj *v1alpha1.UserPass, vaultUserOp cvault.UserPassOperator, vaultToken string) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(obj, userPassFinalizer) {
		err := vaultUserOp.DeleteUserPass(obj.Spec.MountPath, obj.Spec.Name, vaultToken)
		if err != nil {
			return ctrl.Result{RequeueAfter: errorRequeueTime}, err
		}

		patch := client.MergeFrom(obj.DeepCopy())
		controllerutil.RemoveFinalizer(obj, userPassFinalizer)
		if err := r.Patch(ctx, obj, patch); err != nil {
			return ctrl.Result{}, err
		}

	}
	return ctrl.Result{}, nil
}
