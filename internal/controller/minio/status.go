package minio

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch

type StatusReconciler struct {
	utils.DefaultSubReconciler
}

func NewStatusReconciler(client client.Client) *StatusReconciler {
	return &StatusReconciler{
		DefaultSubReconciler: utils.NewDefaultSubReconciler(client),
	}
}

func (r *StatusReconciler) SetupWithManager(ctrlBuilder *builder.Builder) *builder.Builder {
	return ctrlBuilder.Owns(&appsv1.Deployment{})
}

func (r *StatusReconciler) Reconcile(ctx context.Context, minio *v1alpha1.Minio) (ctrl.Result, error) {
	if !minio.DeletionTimestamp.IsZero() {
		// We do not update the status when the resource is already being deleted.
		return ctrl.Result{}, nil
	}

	deployment, err := r.getDeployment(ctx, minio)
	if err != nil {
		return ctrl.Result{}, err
	}

	ready := deployment != nil &&
		deployment.Status.ReadyReplicas > 0 &&
		deployment.Status.Replicas == deployment.Status.ReadyReplicas
	if minio.Status.Ready == ready {
		return ctrl.Result{}, nil
	}

	minio.Status.Ready = ready
	if err := r.GetClient().Status().Update(ctx, minio); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *StatusReconciler) getDeployment(ctx context.Context, minio *v1alpha1.Minio) (*appsv1.Deployment, error) {
	var deployment appsv1.Deployment
	if err := r.GetClient().Get(ctx, client.ObjectKeyFromObject(minio), &deployment); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &deployment, nil
}
