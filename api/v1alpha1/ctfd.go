package v1alpha1

import (
	"slices"

	"github.com/backbone81/ctf-challenge-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CTFdSpec defines the desired state of CTFd.
type CTFdSpec struct {
	// Title is the title for the CTF event.
	// +kubebuilder:validation:Required
	Title string `json:"title"`

	// Description is the description for the CTF event.
	// +kubebuilder:validation:Required
	Description string `json:"description"`

	// UserMode is the user mode for the CTF event.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=teams;users
	// +kubebuilder:default=teams
	UserMode string `json:"userMode"`

	// ChallengeVisibility is the visibility for the challenges.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=public;private;admins
	// +kubebuilder:default=private
	ChallengeVisibility string `json:"challengeVisibility"`

	// AccountVisibility is the visibility for the accounts.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=public;private;admins
	// +kubebuilder:default=private
	AccountVisibility string `json:"accountVisibility"`

	// ScoreVisibility is the visibility for the scores.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=public;private;hidden;admins
	// +kubebuilder:default=private
	ScoreVisibility string `json:"scoreVisibility"`

	// RegistrationVisibility is the visibility for the registration.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=public;private;mlc
	// +kubebuilder:default=private
	RegistrationVisibility string `json:"registrationVisibility"`

	// VerifyEmails specifies if email addresses need to be verified.
	// +kubebuilder:validation:Required
	// +kubebuilder:default=true
	VerifyEmails bool `json:"verifyEmails"`

	// TeamSize specifies the maximum number of members in a team.
	// +kubebuilder:validation:Optional
	TeamSize *int `json:"teamSize"`

	// Theme is the visual theme to use for the website.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=core-beta;core
	// +kubebuilder:default=core-beta
	Theme string `json:"theme"`

	// ThemeColor is the primary color for the theme of the website defined as '#rrggbb'.
	// +kubebuilder:validation:Optional
	ThemeColor *string `json:"themeColor"`

	// Start is the start time of the event.
	// +kubebuilder:validation:Optional
	Start *metav1.Time `json:"start"`

	// End is the end time of the event.
	// +kubebuilder:validation:Optional
	End *metav1.Time `json:"end"`

	// Replicas is the number of replicas to use for the instance.
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas"`

	// Resources specifies resource requests and limits for CPU and memory.
	// +kubebuilder:validation:Optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Redis provides configuration specific to Redis.
	// +kubebuilder:validation:Optional
	Redis RedisSpec `json:"redis"`

	// MariaDB provides configuration specific to MariaDB.
	// +kubebuilder:validation:Optional
	MariaDB MariaDBSpec `json:"mariaDb"`

	// Minio provides configuration specific to Minio.
	// +kubebuilder:validation:Optional
	Minio MinioSpec `json:"minio"`

	// ChallengeNamespace provides the namespace to look for ChallengeDescription resources. Those are then reconciled
	// into the instance. If nil is given, no ChallengeDescriptions are reconciled. If an empty string is given, the
	// same namespace is used.
	// +kubebuilder:validation:Optional
	ChallengeNamespace *string `json:"challengeNamespace"`
}

// CTFdStatus defines the observed state of CTFd.
type CTFdStatus struct {
	// Ready is true when CTFd is up and running.
	// +kubebuilder:validation:Optional
	Ready bool `json:"ready"`

	// ChallengeDescriptions provides information which associates ChallengeDescription resources with database ids
	// of some CTFd instance.
	// +kubebuilder:validation:Optional
	ChallengeDescriptions []ChallengeDescriptionStatus `json:"challengeDescriptions"`
}

func (s *CTFdStatus) GetChallengeDescriptionIndex(challengeDescription v1alpha1.ChallengeDescription) int {
	return slices.IndexFunc(s.ChallengeDescriptions, func(status ChallengeDescriptionStatus) bool {
		return status.Name == challengeDescription.Name && status.Namespace == challengeDescription.Namespace
	})
}

// ChallengeDescriptionStatus provides bookkeeping information about which CTFd challenge id a specific
// ChallengeDescription with the given name in the given namespace was stored as.
type ChallengeDescriptionStatus struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	// +kubebuilder:validation:Optional
	Hints []HintStatus `json:"hints"`
}

// HintStatus provides bookkeeping information about which CTFd hint id a specific hint from the ChallengeDescription
// was stored as.
type HintStatus struct {
	Id    int `json:"id"`    // Id is the database id in CTFd
	Index int `json:"index"` // Index into the slice of hint in the ChallengeDescription
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// CTFd is the Schema for the CTFd.
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
