package mariadb

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

type ServiceReconciler struct {
	utils.DefaultSubReconciler
}

func NewServiceReconciler(client client.Client) *ServiceReconciler {
	return &ServiceReconciler{
		DefaultSubReconciler: utils.NewDefaultSubReconciler(client),
	}
}

func (r *ServiceReconciler) SetupWithManager(ctrlBuilder *builder.Builder) *builder.Builder {
	return ctrlBuilder.Owns(&corev1.Service{})
}

func (r *ServiceReconciler) Reconcile(ctx context.Context, mariadb *v1alpha1.MariaDB) (ctrl.Result, error) {
	currentSpec, err := r.getService(ctx, mariadb)
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredSpec, err := r.getDesiredServiceSpec(mariadb)
	if err != nil {
		return ctrl.Result{}, err
	}

	if currentSpec == nil {
		return r.reconcileOnCreate(ctx, desiredSpec)
	}
	return r.reconcileOnUpdate(ctx, currentSpec, desiredSpec)
}

func (r *ServiceReconciler) reconcileOnCreate(ctx context.Context, desiredSpec *corev1.Service) (ctrl.Result, error) {
	if err := r.GetClient().Create(ctx, desiredSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) reconcileOnUpdate(ctx context.Context, currentSpec *corev1.Service, desiredSpec *corev1.Service) (ctrl.Result, error) {
	if equality.Semantic.DeepDerivative(desiredSpec.Spec, currentSpec.Spec) {
		return ctrl.Result{}, nil
	}
	currentSpec.Spec = desiredSpec.Spec
	if err := r.GetClient().Update(ctx, currentSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) getService(ctx context.Context, mariadb *v1alpha1.MariaDB) (*corev1.Service, error) {
	var service corev1.Service
	if err := r.GetClient().Get(ctx, client.ObjectKeyFromObject(mariadb), &service); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &service, nil
}

func (r *ServiceReconciler) getDesiredServiceSpec(mariadb *v1alpha1.MariaDB) (*corev1.Service, error) {
	result := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mariadb.Name,
			Namespace: mariadb.Namespace,
			Labels:    mariadb.GetDesiredLabels(),
		},
		Spec: corev1.ServiceSpec{
			Selector: mariadb.GetDesiredLabels(),
			Ports: []corev1.ServicePort{
				{
					Name:       "mariadb",
					Port:       3306,
					TargetPort: intstr.FromString("mariadb"),
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(mariadb, &result, r.GetClient().Scheme()); err != nil {
		return nil, err
	}
	return &result, nil
}
