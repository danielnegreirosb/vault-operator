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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppRoleSpec defines the desired state of AppRole
type AppRoleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of AppRole. Edit approle_types.go to remove/update
	// +kubebuilder:validation:Required
	VaultServer *VaultOperatorInstance `json:"vaultOperator"`

	//   - name: external-secret-operator
	// policies:
	//   - external-secret-operator-policy
	// secret_id_ttl: 3600
	// token_ttl: 3600
	// token_max_ttl: 7200
	// export:
	//   namespace: security
	
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:Required
	MountPath string `json:"mount_path,omitempty"`
	// +kubebuilder:default:={"default"}
	// +optional
	Policies []string `json:"policies"`
	// +kubebuilder:default=same
	// +optional
	// +kubebuilder:default=3600
	SecretIDTTL int `json:"secret_id_ttl,omitempty"`
	// Export secret to namespace
	// +optional
	Export *Export `json:"export,omitempty"`

}

type Export struct {
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
}

// AppRoleStatus defines the observed state of AppRole.
type AppRoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the AppRole resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	Synchronized string `json:"synchronized,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
	// +optional
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Synchronized",type=string,JSONPath=".status.synchronized",description="Current AppRole Status"
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=".status.message",description="Status message"
// +kubebuilder:printcolumn:name="Last Update",type=date,JSONPath=".status.lastUpdateTime"

// AppRole is the Schema for the approles API
type AppRole struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of AppRole
	// +required
	Spec AppRoleSpec `json:"spec"`

	// status defines the observed state of AppRole
	// +optional
	Status AppRoleStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// AppRoleList contains a list of AppRole
type AppRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []AppRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppRole{}, &AppRoleList{})
}
