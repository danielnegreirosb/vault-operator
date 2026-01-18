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
	policyFinalizer = "policy.finalizers.ops.community.dev"
)

// PolicyReconciler reconciles a Policy object
type PolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=policies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=policies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vault.ops.community.dev,resources=policies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Policy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/reconcile
func (r *PolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.V(1).Info("Starting Policy Reconciliation")

	obj := &v1alpha1.Policy{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	vaultOpInstance, err := getVaultOpClient(ctx, obj.Spec.VaultServer.Name, 
		obj.Spec.VaultServer.Namespace, req.Namespace, r.Client)
	if err != nil {
		return r.updateStatus(ctx, obj, false,
			fmt.Sprintf("Failed to get vault operator client: %v", err), errorRequeueTime)
	}

	po := cvault.NewPoliciesOperator(vaultOpInstance.Client)

	if !obj.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, obj, po, vaultOpInstance.Token)
	}

	// Ensure finalizer (this modifies spec, not status)
	if !controllerutil.ContainsFinalizer(obj, policyFinalizer) {
		patch := client.MergeFrom(obj.DeepCopy())
		controllerutil.AddFinalizer(obj, policyFinalizer)
		if err := r.Patch(ctx, obj, patch); err != nil {
			return r.updateStatus(ctx, obj, false,
				fmt.Sprintf("Failed to add finalizer: %v", err), errorRequeueTime)
		}
		// Return immediately after adding finalizer to avoid conflicts
		return ctrl.Result{RequeueAfter: 0}, nil
	}

	// Validate Spec
	if err := r.validateSpec(obj); err != nil {
		return r.updateStatus(ctx, obj, false,
			fmt.Sprintf("Failed to sync police %s: %v", obj.Spec.Name, err), errorRequeueTime)
	}

	err = po.CreateOrUpdateAclPolicy(ctx, obj.Spec.Name, obj.Spec.Rules, vaultOpInstance.Token)
	if err != nil {
		return r.updateStatus(ctx, obj, false,
			fmt.Sprintf("Not possible to create/update policy: %s", obj.Spec.Name), errorRequeueTime)
	}

	return r.updateStatus(ctx, obj, true, "Sync", defaultRequeueTime)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vaultv1alpha1.Policy{}).
		Named("policy").
		Complete(r)
}

func (r *PolicyReconciler) validateSpec(obj *v1alpha1.Policy) error {
	if obj.Spec.Name == "" {
		return fmt.Errorf("policy.name cannot be empty")
	}
	if len(obj.Spec.Rules) == 0 {
		return fmt.Errorf("policy.rules needs to contains at least 1 rule")
	}
	return nil
}

func (r *PolicyReconciler) updateStatus(ctx context.Context, obj *v1alpha1.Policy, synced bool, 
	message string, requeueAfter time.Duration) (ctrl.Result, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest := &v1alpha1.Policy{}
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

func (r *PolicyReconciler) handleDeletion(ctx context.Context, obj *v1alpha1.Policy, po *cvault.PoliciesOperator, token string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(obj, policyFinalizer) {
		logger.Info("Cleaning up", "policy", obj.Spec.Name)

		if err := po.DeleteAclPolicy(ctx, obj.Spec.Name, token); err != nil {
			logger.Error(err, "Error cleaning up", "policy", obj.Spec.Name)
			return ctrl.Result{RequeueAfter: errorRequeueTime}, err
		}

		patch := client.MergeFrom(obj.DeepCopy())
		controllerutil.RemoveFinalizer(obj, policyFinalizer)
		if err := r.Patch(ctx, obj, patch); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}
