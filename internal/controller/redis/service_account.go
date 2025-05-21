package redis

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete

type ServiceAccountReconciler struct {
	utils.DefaultSubReconciler
}

func NewServiceAccountReconciler(client client.Client) *ServiceAccountReconciler {
	return &ServiceAccountReconciler{
		DefaultSubReconciler: utils.NewDefaultSubReconciler(client),
	}
}

func (r *ServiceAccountReconciler) SetupWithManager(ctrlBuilder *builder.Builder) *builder.Builder {
	return ctrlBuilder.Owns(&v1.ServiceAccount{})
}

func (r *ServiceAccountReconciler) Reconcile(ctx context.Context, redis *v1alpha1.Redis) (ctrl.Result, error) {
	currentSpec, err := r.getServiceAccount(ctx, redis)
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredSpec, err := r.getDesiredServiceAccountSpec(redis)
	if err != nil {
		return ctrl.Result{}, err
	}

	if currentSpec == nil {
		return r.reconcileOnCreate(ctx, desiredSpec)
	}
	return r.reconcileOnUpdate(ctx, currentSpec, desiredSpec)
}

func (r *ServiceAccountReconciler) reconcileOnCreate(ctx context.Context, desiredSpec *v1.ServiceAccount) (ctrl.Result, error) {
	if err := r.GetClient().Create(ctx, desiredSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ServiceAccountReconciler) reconcileOnUpdate(ctx context.Context, currentSpec *v1.ServiceAccount, desiredSpec *v1.ServiceAccount) (ctrl.Result, error) {
	if equality.Semantic.DeepDerivative(desiredSpec.AutomountServiceAccountToken, currentSpec.AutomountServiceAccountToken) {
		return ctrl.Result{}, nil
	}

	currentSpec.AutomountServiceAccountToken = desiredSpec.AutomountServiceAccountToken
	if err := r.GetClient().Update(ctx, currentSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ServiceAccountReconciler) getServiceAccount(ctx context.Context, redis *v1alpha1.Redis) (*v1.ServiceAccount, error) {
	var serviceAccount v1.ServiceAccount
	if err := r.GetClient().Get(ctx, client.ObjectKeyFromObject(redis), &serviceAccount); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &serviceAccount, nil
}

func (r *ServiceAccountReconciler) getDesiredServiceAccountSpec(redis *v1alpha1.Redis) (*v1.ServiceAccount, error) {
	result := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redis.Name,
			Namespace: redis.Namespace,
			Labels:    redis.GetDesiredLabels(),
		},
		AutomountServiceAccountToken: ptr.To(false),
	}
	if err := controllerutil.SetControllerReference(redis, &result, r.GetClient().Scheme()); err != nil {
		return nil, err
	}
	return &result, nil
}
