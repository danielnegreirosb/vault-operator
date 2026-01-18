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

// VaultServerSpec defines the desired state of VaultServer
type VaultServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// Server contains the vault configuration
	Server VaultServerConfig `json:"server"`

	// Init controls whether to initialize Vault automatically on first start.
	// +kubebuilder:default=true
	// +optional
	Init *bool `json:"init,omitempty"`

	// AutoUnlock controls whether to automatically unseal Vault after initialization.
	// +kubebuilder:default=true
	// +optional
	AutoUnlock *bool `json:"autoUnlock,omitempty"`
}

// VaultServerConfig holds the connection and exposure configuration for Vault.
type VaultServerConfig struct {
	// ServiceName is the name of the Kubernetes Service exposing the Vault server.
	// +kubebuilder:validation:Required
	ServiceName string `json:"serviceName"`

	// Port is the port number for the Vault Server
	// +kubebuilder:default=8200
	// +optional
	Port int32 `json:"port,omitempty"`

	// Namespace is the Kubernetes namespace where the Vault server runs.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// VaultServerStatus defines the observed state of VaultServer.
type VaultServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the VaultServer resource.
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

	// Phase indicates current Vault operation phase
	// +optional
	Phase string `json:"phase,omitempty"`

	// Message provides a human-readable description of the current status.
	// +optional
	Message string `json:"message,omitempty"`

	// LastUpdateTime records the last time the status was updated.
	// +optional
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=".status.phase",description="Current Vault Integration Phase"
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=".status.message",description="Status message"
// +kubebuilder:printcolumn:name="Last Update",type=date,JSONPath=".status.lastUpdateTime"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp"

// VaultServer is the Schema for the vaultservers API
type VaultServer struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of VaultServer
	// +required
	Spec VaultServerSpec `json:"spec"`

	// status defines the observed state of VaultServer
	// +optional
	Status VaultServerStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// VaultServerList contains a list of VaultServer
type VaultServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultServer `json:"items"`
}

// VaultInstance holds the vault Op instance for other obhects reference
type VaultOperatorInstance struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// Namespace is the Kubernetes namespace where the Vault server runs.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

func init() {
	SchemeBuilder.Register(&VaultServer{}, &VaultServerList{})
}
