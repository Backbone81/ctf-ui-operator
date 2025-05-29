package ctfd

import (
	"context"
	"crypto/rand"
	"errors"
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

type AdminSecretReconciler struct {
	utils.DefaultSubReconciler
}

func NewAdminSecretReconciler(client client.Client) *AdminSecretReconciler {
	return &AdminSecretReconciler{
		DefaultSubReconciler: utils.NewDefaultSubReconciler(client),
	}
}

func (r *AdminSecretReconciler) SetupWithManager(ctrlBuilder *builder.Builder) *builder.Builder {
	return ctrlBuilder.Owns(&corev1.Secret{})
}

func (r *AdminSecretReconciler) Reconcile(ctx context.Context, ctfd *v1alpha1.CTFd) (ctrl.Result, error) {
	currentSpec, err := r.getSecret(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredSpec, err := r.getDesiredSecretSpec(ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	if currentSpec == nil {
		return r.reconcileOnCreate(ctx, desiredSpec)
	}
	return r.reconcileOnUpdate(ctx, currentSpec, desiredSpec)
}

func (r *AdminSecretReconciler) reconcileOnCreate(ctx context.Context, desiredSpec *corev1.Secret) (ctrl.Result, error) {
	if err := r.GetClient().Create(ctx, desiredSpec); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *AdminSecretReconciler) reconcileOnUpdate(ctx context.Context, currentSpec *corev1.Secret, desiredSpec *corev1.Secret) (ctrl.Result, error) {
	// This resource is only created, not updated. It would be impossible to recreate the secret with the same
	// credentials.
	return ctrl.Result{}, nil
}

func (r *AdminSecretReconciler) getSecret(ctx context.Context, ctfd *v1alpha1.CTFd) (*corev1.Secret, error) {
	var secret corev1.Secret
	if err := r.GetClient().Get(ctx, client.ObjectKey{
		Name:      AdminSecretName(ctfd),
		Namespace: ctfd.Namespace,
	}, &secret); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &secret, nil
}

func (r *AdminSecretReconciler) getDesiredSecretSpec(ctfd *v1alpha1.CTFd) (*corev1.Secret, error) {
	password, err := r.createRandomPassword()
	if err != nil {
		return nil, err
	}
	result := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AdminSecretName(ctfd),
			Namespace: ctfd.Namespace,
			Labels:    ctfd.GetDesiredLabels(),
		},
		StringData: map[string]string{
			"name":     "admin",
			"email":    "admin@ctfd.internal",
			"password": password,
		},
	}
	if err := controllerutil.SetControllerReference(ctfd, &result, r.GetClient().Scheme()); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *AdminSecretReconciler) randomString(length int) (string, error) {
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

func (r *AdminSecretReconciler) createRandomPassword() (string, error) {
	return r.randomString(32)
}

func AdminSecretName(ctfd *v1alpha1.CTFd) string {
	return ctfd.Name + "-admin"
}

type AdminDetails struct {
	Name     string
	Email    string
	Password string
}

func GetAdminDetails(ctx context.Context, k8sClient client.Client, ctfd *v1alpha1.CTFd) (AdminDetails, error) {
	var secret corev1.Secret
	if err := k8sClient.Get(ctx, client.ObjectKey{
		Name:      AdminSecretName(ctfd),
		Namespace: ctfd.Namespace,
	}, &secret); err != nil {
		return AdminDetails{}, err
	}

	if len(secret.Data["name"]) == 0 {
		return AdminDetails{}, errors.New("user is empty in admin secret")
	}
	if len(secret.Data["email"]) == 0 {
		return AdminDetails{}, errors.New("email is empty in admin secret")
	}
	if len(secret.Data["password"]) == 0 {
		return AdminDetails{}, errors.New("password is empty in admin secret")
	}
	return AdminDetails{
		Name:     string(secret.Data["name"]),
		Email:    string(secret.Data["email"]),
		Password: string(secret.Data["password"]),
	}, nil
}
