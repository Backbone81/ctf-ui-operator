package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RedisSpec defines the desired state of Redis.
type RedisSpec struct{}

// RedisStatus defines the observed state of Redis.
type RedisStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Redis is the Schema for the redis API.
type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisSpec   `json:"spec,omitempty"`
	Status RedisStatus `json:"status,omitempty"`
}

func (r *Redis) GetDesiredLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "redis",
		"app.kubernetes.io/instance": r.Name,
	}
}

// +kubebuilder:object:root=true

// RedisList contains a list of Redis.
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}
