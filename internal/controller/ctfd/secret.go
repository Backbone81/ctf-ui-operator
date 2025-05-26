package ctfd

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

type SecretReconciler struct {
	utils.DefaultSubReconciler
}

func NewSecretReconciler(client client.Client) *SecretReconciler {
	return &SecretReconciler{
		DefaultSubReconciler: utils.NewDefaultSubReconciler(client),
	}
}

func (r *SecretReconciler) SetupWithManager(ctrlBuilder *builder.Builder) *builder.Builder {
	return ctrlBuilder.Owns(&corev1.Secret{})
}

func (r *SecretReconciler) Reconcile(ctx context.Context, ctfd *v1alpha1.CTFd) (ctrl.Result, error) {
	currentSpec, err := r.getSecret(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredSpec, err := r.getDesiredSecretSpec(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	if currentSpec == nil {
		return r.reconcileOnCreate(ctx, desiredSpec)
	}
	return r.reconcileOnUpdate(ctx, currentSpec, desiredSpec)
}

func (r *SecretReconciler) reconcileOnCreate(ctx context.Context, desiredSpec *corev1.Secret) (ctrl.Result, error) {
	if err := r.GetClient().Create(ctx, desiredSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *SecretReconciler) reconcileOnUpdate(ctx context.Context, currentSpec *corev1.Secret, desiredSpec *corev1.Secret) (ctrl.Result, error) {
	// This resource is only created, not updated. It would be impossible to recreate the secret with the same
	// credentials.
	return ctrl.Result{}, nil
}

func (r *SecretReconciler) getSecret(ctx context.Context, ctfd *v1alpha1.CTFd) (*corev1.Secret, error) {
	var secret corev1.Secret
	if err := r.GetClient().Get(ctx, client.ObjectKeyFromObject(ctfd), &secret); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &secret, nil
}

func (r *SecretReconciler) getDesiredSecretSpec(ctx context.Context, ctfd *v1alpha1.CTFd) (*corev1.Secret, error) {
	secretKey, err := r.createRandomSecretKey()
	if err != nil {
		return nil, err
	}
	databaseUrl, err := r.getDatabaseUrl(ctx, ctfd)
	if err != nil {
		return nil, err
	}
	accessKeyId, secretAccessKey, err := r.getMinioCredentials(ctx, ctfd)
	if err != nil {
		return nil, err
	}
	result := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctfd.Name,
			Namespace: ctfd.Namespace,
			Labels:    ctfd.GetDesiredLabels(),
		},
		StringData: map[string]string{
			"SECRET_KEY":   secretKey,
			"DATABASE_URL": databaseUrl,
			"REDIS_URL":    r.getRedisUrl(ctfd),

			"UPLOAD_PROVIDER":         "s3",
			"AWS_ACCESS_KEY_ID":       accessKeyId,
			"AWS_SECRET_ACCESS_KEY":   secretAccessKey,
			"AWS_S3_BUCKET":           ctfd.Name,
			"AWS_S3_ENDPOINT_URL":     r.getMinioUrl(ctfd),
			"AWS_S3_ADDRESSING_STYLE": "path",

			"UPDATE_CHECK": "False",
		},
	}
	if err := controllerutil.SetControllerReference(ctfd, &result, r.GetClient().Scheme()); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *SecretReconciler) randomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}

func (r *SecretReconciler) createRandomSecretKey() (string, error) {
	return r.randomString(128)
}

func (r *SecretReconciler) getDatabaseUrl(ctx context.Context, ctfd *v1alpha1.CTFd) (string, error) {
	hostname := MariaDBName(ctfd)

	var secret corev1.Secret
	if err := r.GetClient().Get(ctx, client.ObjectKey{
		Name:      MariaDBName(ctfd),
		Namespace: ctfd.Namespace,
	}, &secret); err != nil {
		return "", err
	}

	if len(secret.Data["MARIADB_USER"]) == 0 {
		return "", errors.New("MARIADB_USER is empty in MariaDB secret")
	}
	user := string(secret.Data["MARIADB_USER"])

	if len(secret.Data["MARIADB_PASSWORD"]) == 0 {
		return "", errors.New("MARIADB_PASSWORD is empty in MariaDB secret")
	}
	password := string(secret.Data["MARIADB_PASSWORD"])

	if len(secret.Data["MARIADB_DATABASE"]) == 0 {
		return "", errors.New("MARIADB_DATABASE is empty in MariaDB secret")
	}
	database := string(secret.Data["MARIADB_DATABASE"])

	return fmt.Sprintf("mysql+pymysql://%s:%s@%s/%s", user, password, hostname, database), nil
}

func (r *SecretReconciler) getRedisUrl(ctfd *v1alpha1.CTFd) string {
	//nolint:nosprintfhostport // We are not using IPv6 in this URL.
	return fmt.Sprintf("redis://%s:6379", RedisName(ctfd))
}

func (r *SecretReconciler) getMinioUrl(ctfd *v1alpha1.CTFd) string {
	//nolint:nosprintfhostport // We are not using IPv6 in this URL.
	return fmt.Sprintf("http://%s:9000", MinioName(ctfd))
}

func (r *SecretReconciler) getMinioCredentials(ctx context.Context, ctfd *v1alpha1.CTFd) (string, string, error) {
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
