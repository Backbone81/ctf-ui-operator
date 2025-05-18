package ctfd

import (
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=ctfds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=ctfds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=ctfds/finalizers,verbs=update

func NewReconciler(client client.Client, options ...utils.ReconcilerOption[*v1alpha1.CTFd]) *utils.Reconciler[*v1alpha1.CTFd] {
	return utils.NewReconciler[*v1alpha1.CTFd](
		client,
		func() *v1alpha1.CTFd {
			return &v1alpha1.CTFd{}
		},
		options...,
	)
}

// WithDefaultReconcilers returns a reconciler option which enables the default sub-reconcilers.
func WithDefaultReconcilers(recorder record.EventRecorder) utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {}
}
