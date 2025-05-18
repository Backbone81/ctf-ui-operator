package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CTFdSpec defines the desired state of CTFd.
type CTFdSpec struct{}

// CTFdStatus defines the observed state of CTFd.
type CTFdStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// CTFd is the Schema for the apikeys API.
type CTFd struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CTFdSpec   `json:"spec,omitempty"`
	Status CTFdStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CTFdList contains a list of CTFd.
type CTFdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CTFd `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CTFd{}, &CTFdList{})
}
