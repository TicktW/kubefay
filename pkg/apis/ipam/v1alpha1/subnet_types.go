// MIT License Copyright (C) 2022 kubefay@https://github.com/kubefay/kubefay

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
type IPVersionType string

const (
	IPv4 IPVersionType = "IPv4"
	IPv6 IPVersionType = "IPv6"
)

// SubNetSpec defines the desired state of SubNet
type SubNetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	IPVersion IPVersionType `json:"ipVersion"`
	CIDR      string        `json:"cidr,omitempty"`
	// +kubebuilder:validation:Optional
	Gateway string `json:"gateway,omitempty"`
	// +kubebuilder:validation:Optional
	LastReservedIP string `json:"lastReservedIP,omitempty"`
	// +kubebuilder:validation:Optional
	Namespaces []string `json:"namespaces,omitempty"`
	// +kubebuilder:validation:Optional
	ExternalIPs []string `json:"externalIPs,omitempty"`
	// +kubebuilder:validation:Optional
	UsedPool map[string]string `json:"usedPool,omitempty"`
	// +kubebuilder:validation:Optional
	UnusedPool []string `json:"unusedPool,omitempty"`
	// +kubebuilder:validation:Optional
	DNS map[string]string `json:"dns,omitempty"`
}

// SubNetStatus defines the observed state of SubNet
type SubNetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	PoolStatus string `json:"poolStatus,omitempty"`
	IPAMEvent  string `json:"ipamEvent,omitempty"`
}

// +kubebuilder:object:root=true
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SubNet is the Schema for the subnets API
type SubNet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubNetSpec   `json:"spec,omitempty"`
	Status SubNetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SubNetList contains a list of SubNet
type SubNetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SubNet `json:"items"`
}
