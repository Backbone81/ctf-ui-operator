package minio

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

type PersistentVolumeClaimReconciler struct {
	utils.DefaultSubReconciler
}

func NewPersistentVolumeClaimReconciler(client client.Client) *PersistentVolumeClaimReconciler {
	return &PersistentVolumeClaimReconciler{
		DefaultSubReconciler: utils.NewDefaultSubReconciler(client),
	}
}

func (r *PersistentVolumeClaimReconciler) SetupWithManager(ctrlBuilder *builder.Builder) *builder.Builder {
	return ctrlBuilder.Owns(&corev1.PersistentVolumeClaim{})
}

func (r *PersistentVolumeClaimReconciler) Reconcile(ctx context.Context, minio *v1alpha1.Minio) (ctrl.Result, error) {
	currentSpec, err := r.getPersistentVolumeClaim(ctx, minio)
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredSpec, err := r.getDesiredPersistentVolumeClaimSpec(minio)
	if err != nil {
		return ctrl.Result{}, err
	}

	if currentSpec == nil {
		return r.reconcileOnCreate(ctx, desiredSpec)
	}
	return r.reconcileOnUpdate(ctx, currentSpec, desiredSpec)
}

func (r *PersistentVolumeClaimReconciler) reconcileOnCreate(ctx context.Context, desiredSpec *corev1.PersistentVolumeClaim) (ctrl.Result, error) {
	if err := r.GetClient().Create(ctx, desiredSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *PersistentVolumeClaimReconciler) reconcileOnUpdate(ctx context.Context, currentSpec *corev1.PersistentVolumeClaim, desiredSpec *corev1.PersistentVolumeClaim) (ctrl.Result, error) {
	if equality.Semantic.DeepDerivative(desiredSpec.Spec, currentSpec.Spec) {
		return ctrl.Result{}, nil
	}

	currentSpec.Spec = desiredSpec.Spec
	if err := r.GetClient().Update(ctx, currentSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *PersistentVolumeClaimReconciler) getPersistentVolumeClaim(ctx context.Context, minio *v1alpha1.Minio) (*corev1.PersistentVolumeClaim, error) {
	var persistentVolumeClaim corev1.PersistentVolumeClaim
	if err := r.GetClient().Get(ctx, client.ObjectKeyFromObject(minio), &persistentVolumeClaim); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &persistentVolumeClaim, nil
}

func (r *PersistentVolumeClaimReconciler) getDesiredPersistentVolumeClaimSpec(minio *v1alpha1.Minio) (*corev1.PersistentVolumeClaim, error) {
	storageSize, err := resource.ParseQuantity("128Mi")
	if err != nil {
		return nil, err
	}
	result := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      minio.Name,
			Namespace: minio.Namespace,
			Labels:    minio.GetDesiredLabels(),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageSize,
				},
			},
		},
	}
	if minio.Spec.PersistentVolumeClaim != nil {
		result.Spec = *minio.Spec.PersistentVolumeClaim
	}
	if err := controllerutil.SetControllerReference(minio, &result, r.GetClient().Scheme()); err != nil {
		return nil, err
	}
	return &result, nil
}
