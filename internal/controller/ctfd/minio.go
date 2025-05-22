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

// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=minios,verbs=get;list;watch;create;update;patch;delete

type MinioReconciler struct {
	utils.DefaultSubReconciler
}

func NewMinioReconciler(client client.Client) *MinioReconciler {
	return &MinioReconciler{
		DefaultSubReconciler: utils.NewDefaultSubReconciler(client),
	}
}

func (r *MinioReconciler) SetupWithManager(ctrlBuilder *builder.Builder) *builder.Builder {
	return ctrlBuilder.Owns(&v1alpha1.Minio{})
}

func (r *MinioReconciler) Reconcile(ctx context.Context, ctfd *v1alpha1.CTFd) (ctrl.Result, error) {
	currentSpec, err := r.getMinio(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredSpec, err := r.getDesiredMinioSpec(ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	if currentSpec == nil {
		return r.reconcileOnCreate(ctx, desiredSpec)
	}
	return r.reconcileOnUpdate(ctx, currentSpec, desiredSpec)
}

func (r *MinioReconciler) reconcileOnCreate(ctx context.Context, desiredSpec *v1alpha1.Minio) (ctrl.Result, error) {
	if err := r.GetClient().Create(ctx, desiredSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *MinioReconciler) reconcileOnUpdate(ctx context.Context, currentSpec *v1alpha1.Minio, desiredSpec *v1alpha1.Minio) (ctrl.Result, error) {
	if equality.Semantic.DeepDerivative(desiredSpec.Spec, currentSpec.Spec) {
		return ctrl.Result{}, nil
	}

	currentSpec.Spec = desiredSpec.Spec
	if err := r.GetClient().Update(ctx, currentSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *MinioReconciler) getMinio(ctx context.Context, ctfd *v1alpha1.CTFd) (*v1alpha1.Minio, error) {
	var minio v1alpha1.Minio
	if err := r.GetClient().Get(ctx, client.ObjectKey{
		Name:      MinioName(ctfd),
		Namespace: ctfd.Namespace,
	}, &minio); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &minio, nil
}

func (r *MinioReconciler) getDesiredMinioSpec(ctfd *v1alpha1.CTFd) (*v1alpha1.Minio, error) {
	result := v1alpha1.Minio{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MinioName(ctfd),
			Namespace: ctfd.Namespace,
			Labels:    ctfd.GetDesiredLabels(),
		},
		Spec: ctfd.Spec.Minio,
	}
	if err := controllerutil.SetControllerReference(ctfd, &result, r.GetClient().Scheme()); err != nil {
		return nil, err
	}
	return &result, nil
}

func MinioName(ctfd *v1alpha1.CTFd) string {
	return ctfd.Name + "-minio"
}
