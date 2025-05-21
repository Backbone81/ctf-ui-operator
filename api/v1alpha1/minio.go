package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MinioSpec defines the desired state of Minio.
type MinioSpec struct {
	// Resources specifies resource requests and limits for CPU and memory.
	// +kubebuilder:validation:Optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// PersistentVolumeClaim is the storage to allocate for the Minio instance.
	// +kubebuilder:validation:Optional
	PersistentVolumeClaim *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
}

// MinioStatus defines the observed state of Minio.
type MinioStatus struct {
	// Ready is true when Minio is up and running.
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Minio is the Schema for the Minio API.
type Minio struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinioSpec   `json:"spec,omitempty"`
	Status MinioStatus `json:"status,omitempty"`
}

func (r *Minio) GetDesiredLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "Minio",
		"app.kubernetes.io/instance": r.Name,
	}
}

// +kubebuilder:object:root=true

// MinioList contains a list of Minio.
type MinioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Minio `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Minio{}, &MinioList{})
}
