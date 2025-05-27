package ctfd

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

//nolint:cyclop // The ready status is dependent on several other resources.
func (r *StatusReconciler) Reconcile(ctx context.Context, ctfd *v1alpha1.CTFd) (ctrl.Result, error) {
	if !ctfd.DeletionTimestamp.IsZero() {
		// We do not update the status when the resource is already being deleted.
		return ctrl.Result{}, nil
	}

	redis, err := r.getRedis(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	mariaDB, err := r.getMariaDB(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	minio, err := r.getMinio(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	deployment, err := r.getDeployment(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	ready := redis != nil && redis.Status.Ready &&
		mariaDB != nil && mariaDB.Status.Ready &&
		minio != nil && minio.Status.Ready &&
		deployment != nil &&
		deployment.Status.ReadyReplicas > 0 &&
		deployment.Status.Replicas == deployment.Status.ReadyReplicas
	if ctfd.Status.Ready == ready {
		return ctrl.Result{}, nil
	}

	ctfd.Status.Ready = ready
	if err := r.GetClient().Status().Update(ctx, ctfd); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *StatusReconciler) getDeployment(ctx context.Context, ctfd *v1alpha1.CTFd) (*appsv1.Deployment, error) {
	var deployment appsv1.Deployment
	if err := r.GetClient().Get(ctx, client.ObjectKeyFromObject(ctfd), &deployment); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &deployment, nil
}

func (r *StatusReconciler) getRedis(ctx context.Context, ctfd *v1alpha1.CTFd) (*v1alpha1.Redis, error) {
	var redis v1alpha1.Redis
	if err := r.GetClient().Get(ctx, client.ObjectKey{
		Namespace: ctfd.Namespace,
		Name:      RedisName(ctfd),
	}, &redis); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &redis, nil
}

func (r *StatusReconciler) getMariaDB(ctx context.Context, ctfd *v1alpha1.CTFd) (*v1alpha1.MariaDB, error) {
	var mariaDB v1alpha1.MariaDB
	if err := r.GetClient().Get(ctx, client.ObjectKey{
		Namespace: ctfd.Namespace,
		Name:      MariaDBName(ctfd),
	}, &mariaDB); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &mariaDB, nil
}

func (r *StatusReconciler) getMinio(ctx context.Context, ctfd *v1alpha1.CTFd) (*v1alpha1.Minio, error) {
	var minio v1alpha1.Minio
	if err := r.GetClient().Get(ctx, client.ObjectKey{
		Namespace: ctfd.Namespace,
		Name:      MinioName(ctfd),
	}, &minio); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &minio, nil
}
