package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CTFdSpec defines the desired state of CTFd.
type CTFdSpec struct {
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas"`

	// Resources specifies resource requests and limits for CPU and memory.
	// +kubebuilder:validation:Optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Redis provides configuration specific to Redis.
	Redis RedisSpec `json:"redis"`

	// MariaDB provides configuration specific to MariaDB.
	MariaDB MariaDBSpec `json:"mariaDb"`

	// Minio provides configuration specific to Minio.
	Minio MinioSpec `json:"minio"`
}

// CTFdStatus defines the observed state of CTFd.
type CTFdStatus struct {
	// Ready is true when CTFd is up and running.
	Ready bool `json:"ready"`
}

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

func (r *CTFd) GetDesiredLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "CTFd",
		"app.kubernetes.io/instance": r.Name,
	}
}

func (r *CTFd) GetReplicas() int32 {
	if r.Spec.Replicas == nil {
		return 1
	}
	return *r.Spec.Replicas
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
