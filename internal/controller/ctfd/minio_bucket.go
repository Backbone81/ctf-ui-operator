package ctfd

import (
	"context"
	"errors"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=ui.ctf.backbone81,resources=minios,verbs=get;list;watch;create;update;patch;delete

// MinioBucketReconciler is responsible for creating the bucket for the CTFd instance.
type MinioBucketReconciler struct {
	utils.DefaultSubReconciler
	minioEndpoint MinioEndpointStrategy
}

func NewMinioBucketReconciler(client client.Client, options ...SubReconcilerOption) *MinioBucketReconciler {
	result := &MinioBucketReconciler{
		DefaultSubReconciler: utils.NewDefaultSubReconciler(client),
	}
	for _, option := range options {
		option(result)
	}

	if result.minioEndpoint == nil {
		panic("Minio endpoint strategy required")
	}
	return result
}

func (r *MinioBucketReconciler) SetupWithManager(ctrlBuilder *builder.Builder) *builder.Builder {
	return ctrlBuilder.Owns(&v1alpha1.Minio{})
}

func (r *MinioBucketReconciler) Reconcile(ctx context.Context, ctfd *v1alpha1.CTFd) (ctrl.Result, error) {
	currentSpec, err := r.getMinio(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}
	if currentSpec == nil {
		// The minio instance is not available yet. We will get triggered later again.
		ctrl.LoggerFrom(ctx).V(1).Info("Minio not found, skipping MinioBucketReconciler.")
		return ctrl.Result{}, nil
	}

	if !currentSpec.Status.Ready {
		// The Minio instance is not ready. We try again later when the instance is up and running. The next reconcile
		// will be triggered when the status changes.
		ctrl.LoggerFrom(ctx).V(1).Info("Minio is not ready, skipping MinioBucketReconciler.")
		return ctrl.Result{}, nil
	}

	accessKeyID, secretAccessKey, err := r.getMinioCredentials(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}
	minioEndpoint, err := r.minioEndpoint.GetEndpoint(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	exists, err := minioClient.BucketExists(ctx, ctfd.Name)
	if err != nil {
		return ctrl.Result{}, err
	}
	if exists {
		// We can exit early, as the bucket already exists.
		return ctrl.Result{}, nil
	}

	ctrl.LoggerFrom(ctx).Info("Creating Minio bucket", "bucket", ctfd.Name)
	if err := minioClient.MakeBucket(ctx, ctfd.Name, minio.MakeBucketOptions{}); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *MinioBucketReconciler) getMinio(ctx context.Context, ctfd *v1alpha1.CTFd) (*v1alpha1.Minio, error) {
	var minio v1alpha1.Minio
	if err := r.GetClient().Get(ctx, client.ObjectKey{
		Name:      MinioName(ctfd),
		Namespace: ctfd.Namespace,
	}, &minio); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &minio, nil
}

func (r *MinioBucketReconciler) getMinioCredentials(ctx context.Context, ctfd *v1alpha1.CTFd) (string, string, error) {
	var secret corev1.Secret
	if err := r.GetClient().Get(ctx, client.ObjectKey{
		Name:      MinioName(ctfd),
		Namespace: ctfd.Namespace,
	}, &secret); err != nil {
		return "", "", err
	}

	if len(secret.Data["MINIO_ROOT_USER"]) == 0 {
		return "", "", errors.New("MINIO_ROOT_USER is empty in Minio secret")
	}
	accessKeyId := string(secret.Data["MINIO_ROOT_USER"])

	if len(secret.Data["MINIO_ROOT_PASSWORD"]) == 0 {
		return "", "", errors.New("MINIO_ROOT_PASSWORD is empty in Minio secret")
	}
	secretAccessKey := string(secret.Data["MINIO_ROOT_PASSWORD"])

	return accessKeyId, secretAccessKey, nil
}

func (r *MinioBucketReconciler) SetMinioEndpoint(endpoint MinioEndpointStrategy) {
	r.minioEndpoint = endpoint
}
