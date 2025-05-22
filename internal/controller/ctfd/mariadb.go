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

// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=mariadbs,verbs=get;list;watch;create;update;patch;delete

type MariaDBReconciler struct {
	utils.DefaultSubReconciler
}

func NewMariaDBReconciler(client client.Client) *MariaDBReconciler {
	return &MariaDBReconciler{
		DefaultSubReconciler: utils.NewDefaultSubReconciler(client),
	}
}

func (r *MariaDBReconciler) SetupWithManager(ctrlBuilder *builder.Builder) *builder.Builder {
	return ctrlBuilder.Owns(&v1alpha1.MariaDB{})
}

func (r *MariaDBReconciler) Reconcile(ctx context.Context, ctfd *v1alpha1.CTFd) (ctrl.Result, error) {
	currentSpec, err := r.getMariaDB(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredSpec, err := r.getDesiredMariaDBSpec(ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	if currentSpec == nil {
		return r.reconcileOnCreate(ctx, desiredSpec)
	}
	return r.reconcileOnUpdate(ctx, currentSpec, desiredSpec)
}

func (r *MariaDBReconciler) reconcileOnCreate(ctx context.Context, desiredSpec *v1alpha1.MariaDB) (ctrl.Result, error) {
	if err := r.GetClient().Create(ctx, desiredSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *MariaDBReconciler) reconcileOnUpdate(ctx context.Context, currentSpec *v1alpha1.MariaDB, desiredSpec *v1alpha1.MariaDB) (ctrl.Result, error) {
	if equality.Semantic.DeepDerivative(desiredSpec.Spec, currentSpec.Spec) {
		return ctrl.Result{}, nil
	}

	currentSpec.Spec = desiredSpec.Spec
	if err := r.GetClient().Update(ctx, currentSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *MariaDBReconciler) getMariaDB(ctx context.Context, ctfd *v1alpha1.CTFd) (*v1alpha1.MariaDB, error) {
	var mariadb v1alpha1.MariaDB
	if err := r.GetClient().Get(ctx, client.ObjectKey{
		Name:      MariaDBName(ctfd),
		Namespace: ctfd.Namespace,
	}, &mariadb); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &mariadb, nil
}

func (r *MariaDBReconciler) getDesiredMariaDBSpec(ctfd *v1alpha1.CTFd) (*v1alpha1.MariaDB, error) {
	result := v1alpha1.MariaDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MariaDBName(ctfd),
			Namespace: ctfd.Namespace,
			Labels:    ctfd.GetDesiredLabels(),
		},
		Spec: ctfd.Spec.MariaDB,
	}
	if err := controllerutil.SetControllerReference(ctfd, &result, r.GetClient().Scheme()); err != nil {
		return nil, err
	}
	return &result, nil
}

func MariaDBName(ctfd *v1alpha1.CTFd) string {
	return ctfd.Name + "-mariadb"
}
