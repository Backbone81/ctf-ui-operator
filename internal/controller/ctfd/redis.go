//nolint:dupl // The reconcilers for Redis, MariaDB and Minio are intentionally similar.
package ctfd

import (
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=redis,verbs=get;list;watch;create;update;patch;delete

type RedisReconciler struct {
	utils.DefaultSubReconciler
}

func NewRedisReconciler(client client.Client) *RedisReconciler {
	return &RedisReconciler{
		DefaultSubReconciler: utils.NewDefaultSubReconciler(client),
	}
}

func (r *RedisReconciler) SetupWithManager(ctrlBuilder *builder.Builder) *builder.Builder {
	return ctrlBuilder.Owns(&v1alpha1.Redis{})
}

func (r *RedisReconciler) Reconcile(ctx context.Context, ctfd *v1alpha1.CTFd) (ctrl.Result, error) {
	currentSpec, err := r.getRedis(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredSpec, err := r.getDesiredRedisSpec(ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	if currentSpec == nil {
		return r.reconcileOnCreate(ctx, desiredSpec)
	}
	return r.reconcileOnUpdate(ctx, currentSpec, desiredSpec)
}

func (r *RedisReconciler) reconcileOnCreate(ctx context.Context, desiredSpec *v1alpha1.Redis) (ctrl.Result, error) {
	if err := r.GetClient().Create(ctx, desiredSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *RedisReconciler) reconcileOnUpdate(ctx context.Context, currentSpec *v1alpha1.Redis, desiredSpec *v1alpha1.Redis) (ctrl.Result, error) {
	if equality.Semantic.DeepDerivative(desiredSpec.Spec, currentSpec.Spec) {
		return ctrl.Result{}, nil
	}

	currentSpec.Spec = desiredSpec.Spec
	if err := r.GetClient().Update(ctx, currentSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *RedisReconciler) getRedis(ctx context.Context, ctfd *v1alpha1.CTFd) (*v1alpha1.Redis, error) {
	var redis v1alpha1.Redis
	if err := r.GetClient().Get(ctx, client.ObjectKey{
		Name:      RedisName(ctfd),
		Namespace: ctfd.Namespace,
	}, &redis); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &redis, nil
}

func (r *RedisReconciler) getDesiredRedisSpec(ctfd *v1alpha1.CTFd) (*v1alpha1.Redis, error) {
	result := v1alpha1.Redis{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RedisName(ctfd),
			Namespace: ctfd.Namespace,
			Labels:    ctfd.GetDesiredLabels(),
		},
		Spec: ctfd.Spec.Redis,
	}
	if err := controllerutil.SetControllerReference(ctfd, &result, r.GetClient().Scheme()); err != nil {
		return nil, err
	}
	return &result, nil
}

func RedisName(ctfd *v1alpha1.CTFd) string {
	return ctfd.Name + "-redis"
}
