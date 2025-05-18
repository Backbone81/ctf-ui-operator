package redis

import (
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=redis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=redis/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=redis/finalizers,verbs=update

func NewReconciler(client client.Client, options ...utils.ReconcilerOption[*v1alpha1.Redis]) *utils.Reconciler[*v1alpha1.Redis] {
	return utils.NewReconciler[*v1alpha1.Redis](
		client,
		func() *v1alpha1.Redis {
			return &v1alpha1.Redis{}
		},
		options...,
	)
}

// WithDefaultReconcilers returns a reconciler option which enables the default sub-reconcilers.
func WithDefaultReconcilers(recorder record.EventRecorder) utils.ReconcilerOption[*v1alpha1.Redis] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Redis]) {
		WithServiceAccountReconciler()(reconciler)
	}
}

func WithServiceAccountReconciler() utils.ReconcilerOption[*v1alpha1.Redis] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Redis]) {
		reconciler.AppendSubReconciler(NewServiceAccountReconciler(reconciler.GetClient()))
	}
}
