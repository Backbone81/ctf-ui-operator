package ctfd

import (
	"context"

	corev1 "k8s.io/api/core/v1"
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
	return ctrlBuilder.Owns(&corev1.ServiceAccount{})
}

func (r *ServiceAccountReconciler) Reconcile(ctx context.Context, ctfd *v1alpha1.CTFd) (ctrl.Result, error) {
	currentSpec, err := r.getServiceAccount(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredSpec, err := r.getDesiredServiceAccountSpec(ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	if currentSpec == nil {
		return r.reconcileOnCreate(ctx, desiredSpec)
	}
	return r.reconcileOnUpdate(ctx, currentSpec, desiredSpec)
}

func (r *ServiceAccountReconciler) reconcileOnCreate(ctx context.Context, desiredSpec *corev1.ServiceAccount) (ctrl.Result, error) {
	if err := r.GetClient().Create(ctx, desiredSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ServiceAccountReconciler) reconcileOnUpdate(ctx context.Context, currentSpec *corev1.ServiceAccount, desiredSpec *corev1.ServiceAccount) (ctrl.Result, error) {
	if equality.Semantic.DeepDerivative(desiredSpec.AutomountServiceAccountToken, currentSpec.AutomountServiceAccountToken) {
		return ctrl.Result{}, nil
	}

	currentSpec.AutomountServiceAccountToken = desiredSpec.AutomountServiceAccountToken
	if err := r.GetClient().Update(ctx, currentSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ServiceAccountReconciler) getServiceAccount(ctx context.Context, ctfd *v1alpha1.CTFd) (*corev1.ServiceAccount, error) {
	var serviceAccount corev1.ServiceAccount
	if err := r.GetClient().Get(ctx, client.ObjectKeyFromObject(ctfd), &serviceAccount); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &serviceAccount, nil
}

func (r *ServiceAccountReconciler) getDesiredServiceAccountSpec(ctfd *v1alpha1.CTFd) (*corev1.ServiceAccount, error) {
	result := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctfd.Name,
			Namespace: ctfd.Namespace,
			Labels:    ctfd.GetDesiredLabels(),
		},
		AutomountServiceAccountToken: ptr.To(false),
	}
	if err := controllerutil.SetControllerReference(ctfd, &result, r.GetClient().Scheme()); err != nil {
		return nil, err
	}
	return &result, nil
}
