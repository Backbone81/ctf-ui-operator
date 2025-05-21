package mariadb

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=mariadbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=mariadbs/finalizers,verbs=update
// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=mariadbs/status,verbs=get;update;patch

func NewReconciler(client client.Client, options ...utils.ReconcilerOption[*v1alpha1.MariaDB]) *utils.Reconciler[*v1alpha1.MariaDB] {
	return utils.NewReconciler[*v1alpha1.MariaDB](
		client,
		func() *v1alpha1.MariaDB {
			return &v1alpha1.MariaDB{}
		},
		options...,
	)
}

// WithDefaultReconcilers returns a reconciler option which enables the default sub-reconcilers.
func WithDefaultReconcilers() utils.ReconcilerOption[*v1alpha1.MariaDB] {
	return func(reconciler *utils.Reconciler[*v1alpha1.MariaDB]) {
		WithStatusReconciler()(reconciler)
		WithServiceAccountReconciler()(reconciler)
		WithServiceReconciler()(reconciler)
		WithPersistentVolumeClaimReconciler()(reconciler)
		WithSecretReconciler()(reconciler)
		WithDeploymentReconciler()(reconciler)
	}
}

func WithDeploymentReconciler() utils.ReconcilerOption[*v1alpha1.MariaDB] {
	return func(reconciler *utils.Reconciler[*v1alpha1.MariaDB]) {
		reconciler.AppendSubReconciler(NewDeploymentReconciler(reconciler.GetClient()))
	}
}

func WithPersistentVolumeClaimReconciler() utils.ReconcilerOption[*v1alpha1.MariaDB] {
	return func(reconciler *utils.Reconciler[*v1alpha1.MariaDB]) {
		reconciler.AppendSubReconciler(NewPersistentVolumeClaimReconciler(reconciler.GetClient()))
	}
}

func WithServiceReconciler() utils.ReconcilerOption[*v1alpha1.MariaDB] {
	return func(reconciler *utils.Reconciler[*v1alpha1.MariaDB]) {
		reconciler.AppendSubReconciler(NewServiceReconciler(reconciler.GetClient()))
	}
}

func WithSecretReconciler() utils.ReconcilerOption[*v1alpha1.MariaDB] {
	return func(reconciler *utils.Reconciler[*v1alpha1.MariaDB]) {
		reconciler.AppendSubReconciler(NewSecretReconciler(reconciler.GetClient()))
	}
}

func WithServiceAccountReconciler() utils.ReconcilerOption[*v1alpha1.MariaDB] {
	return func(reconciler *utils.Reconciler[*v1alpha1.MariaDB]) {
		reconciler.AppendSubReconciler(NewServiceAccountReconciler(reconciler.GetClient()))
	}
}

func WithStatusReconciler() utils.ReconcilerOption[*v1alpha1.MariaDB] {
	return func(reconciler *utils.Reconciler[*v1alpha1.MariaDB]) {
		reconciler.AppendSubReconciler(NewStatusReconciler(reconciler.GetClient()))
	}
}
