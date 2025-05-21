package minio

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=minios,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=minios/finalizers,verbs=update
// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=minios/status,verbs=get;update;patch

func NewReconciler(client client.Client, options ...utils.ReconcilerOption[*v1alpha1.Minio]) *utils.Reconciler[*v1alpha1.Minio] {
	return utils.NewReconciler[*v1alpha1.Minio](
		client,
		func() *v1alpha1.Minio {
			return &v1alpha1.Minio{}
		},
		options...,
	)
}

// WithDefaultReconcilers returns a reconciler option which enables the default sub-reconcilers.
func WithDefaultReconcilers() utils.ReconcilerOption[*v1alpha1.Minio] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Minio]) {
		WithStatusReconciler()(reconciler)
		WithServiceAccountReconciler()(reconciler)
		WithServiceReconciler()(reconciler)
		WithPersistentVolumeClaimReconciler()(reconciler)
		WithSecretReconciler()(reconciler)
		WithDeploymentReconciler()(reconciler)
	}
}

func WithDeploymentReconciler() utils.ReconcilerOption[*v1alpha1.Minio] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Minio]) {
		reconciler.AppendSubReconciler(NewDeploymentReconciler(reconciler.GetClient()))
	}
}

func WithPersistentVolumeClaimReconciler() utils.ReconcilerOption[*v1alpha1.Minio] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Minio]) {
		reconciler.AppendSubReconciler(NewPersistentVolumeClaimReconciler(reconciler.GetClient()))
	}
}

func WithServiceReconciler() utils.ReconcilerOption[*v1alpha1.Minio] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Minio]) {
		reconciler.AppendSubReconciler(NewServiceReconciler(reconciler.GetClient()))
	}
}

func WithSecretReconciler() utils.ReconcilerOption[*v1alpha1.Minio] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Minio]) {
		reconciler.AppendSubReconciler(NewSecretReconciler(reconciler.GetClient()))
	}
}

func WithServiceAccountReconciler() utils.ReconcilerOption[*v1alpha1.Minio] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Minio]) {
		reconciler.AppendSubReconciler(NewServiceAccountReconciler(reconciler.GetClient()))
	}
}

func WithStatusReconciler() utils.ReconcilerOption[*v1alpha1.Minio] {
	return func(reconciler *utils.Reconciler[*v1alpha1.Minio]) {
		reconciler.AppendSubReconciler(NewStatusReconciler(reconciler.GetClient()))
	}
}
