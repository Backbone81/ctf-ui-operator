package ctfd

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=ctfds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=ctfds/finalizers,verbs=update
// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=ctfds/status,verbs=get;update;patch

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
func WithDefaultReconcilers() utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		WithStatusReconciler()(reconciler)

		WithMariaDBReconciler()(reconciler)
		WithRedisReconciler()(reconciler)
		WithMinioReconciler()(reconciler)
		WithMinioBucketReconciler(WithMinioAutodetectEndpoint())(reconciler)

		WithServiceAccountReconciler()(reconciler)
		WithServiceReconciler()(reconciler)
		WithSecretReconciler()(reconciler)
		WithDeploymentReconciler()(reconciler)

		WithAdminSecretReconciler()(reconciler)
		WithSetupReconciler(WithCTFdAutodetectEndpoint())(reconciler)
		WithAccessTokenReconciler(WithCTFdAutodetectEndpoint())(reconciler)
		WithChallengeDescriptionReconciler(WithCTFdAutodetectEndpoint())(reconciler)
	}
}

func WithAccessTokenReconciler(options ...SubReconcilerOption) utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		reconciler.AppendSubReconciler(NewAccessTokenReconciler(reconciler.GetClient(), options...))
	}
}

func WithAdminSecretReconciler() utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		reconciler.AppendSubReconciler(NewAdminSecretReconciler(reconciler.GetClient()))
	}
}

func WithChallengeDescriptionReconciler(options ...SubReconcilerOption) utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		reconciler.AppendSubReconciler(NewChallengeDescriptionReconciler(reconciler.GetClient(), options...))
	}
}

func WithDeploymentReconciler() utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		reconciler.AppendSubReconciler(NewDeploymentReconciler(reconciler.GetClient()))
	}
}

func WithMariaDBReconciler() utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		reconciler.AppendSubReconciler(NewMariaDBReconciler(reconciler.GetClient()))
	}
}

func WithMinioReconciler() utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		reconciler.AppendSubReconciler(NewMinioReconciler(reconciler.GetClient()))
	}
}

func WithMinioBucketReconciler(options ...SubReconcilerOption) utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		reconciler.AppendSubReconciler(NewMinioBucketReconciler(reconciler.GetClient(), options...))
	}
}

func WithRedisReconciler() utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		reconciler.AppendSubReconciler(NewRedisReconciler(reconciler.GetClient()))
	}
}

func WithSecretReconciler() utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		reconciler.AppendSubReconciler(NewSecretReconciler(reconciler.GetClient()))
	}
}

func WithServiceReconciler() utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		reconciler.AppendSubReconciler(NewServiceReconciler(reconciler.GetClient()))
	}
}

func WithServiceAccountReconciler() utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		reconciler.AppendSubReconciler(NewServiceAccountReconciler(reconciler.GetClient()))
	}
}

func WithSetupReconciler(options ...SubReconcilerOption) utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		reconciler.AppendSubReconciler(NewSetupReconciler(reconciler.GetClient(), options...))
	}
}

func WithStatusReconciler() utils.ReconcilerOption[*v1alpha1.CTFd] {
	return func(reconciler *utils.Reconciler[*v1alpha1.CTFd]) {
		reconciler.AppendSubReconciler(NewStatusReconciler(reconciler.GetClient()))
	}
}
