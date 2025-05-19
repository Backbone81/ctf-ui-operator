package redis

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=redis,verbs=get;list;watch;create;update;patch;delete
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
func WithDefaultReconcilers() utils.ReconcilerOption[*v1alpha1.Redis] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Redis]) {
		WithServiceAccountReconciler()(reconciler)
		WithServiceReconciler()(reconciler)
		WithPersistentVolumeClaimReconciler()(reconciler)
		WithDeploymentReconciler()(reconciler)
		WithStatusReconciler()(reconciler)
	}
}

func WithDeploymentReconciler() utils.ReconcilerOption[*v1alpha1.Redis] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Redis]) {
		reconciler.AppendSubReconciler(NewDeploymentReconciler(reconciler.GetClient()))
	}
}

func WithPersistentVolumeClaimReconciler() utils.ReconcilerOption[*v1alpha1.Redis] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Redis]) {
		reconciler.AppendSubReconciler(NewPersistentVolumeClaimReconciler(reconciler.GetClient()))
	}
}

func WithServiceReconciler() utils.ReconcilerOption[*v1alpha1.Redis] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Redis]) {
		reconciler.AppendSubReconciler(NewServiceReconciler(reconciler.GetClient()))
	}
}

func WithServiceAccountReconciler() utils.ReconcilerOption[*v1alpha1.Redis] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Redis]) {
		reconciler.AppendSubReconciler(NewServiceAccountReconciler(reconciler.GetClient()))
	}
}

func WithStatusReconciler() utils.ReconcilerOption[*v1alpha1.Redis] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Redis]) {
		reconciler.AppendSubReconciler(NewStatusReconciler(reconciler.GetClient()))
	}
}
