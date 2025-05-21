package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MariaDBSpec defines the desired state of MariaDB.
type MariaDBSpec struct {
	// Resources specifies resource requests and limits for CPU and memory.
	// +kubebuilder:validation:Optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// PersistentVolumeClaim is the storage to allocate for the MariaDB instance.
	// +kubebuilder:validation:Optional
	PersistentVolumeClaim *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
}

// MariaDBStatus defines the observed state of MariaDB.
type MariaDBStatus struct {
	// Ready is true when MariaDB is up and running.
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// MariaDB is the Schema for the MariaDB API.
type MariaDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MariaDBSpec   `json:"spec,omitempty"`
	Status MariaDBStatus `json:"status,omitempty"`
}

func (r *MariaDB) GetDesiredLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "MariaDB",
		"app.kubernetes.io/instance": r.Name,
	}
}

// +kubebuilder:object:root=true

// MariaDBList contains a list of MariaDB.
type MariaDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MariaDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MariaDB{}, &MariaDBList{})
}
