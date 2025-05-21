package mariadb

import (
	"context"
	"crypto/rand"
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

func (r *SecretReconciler) Reconcile(ctx context.Context, mariadb *v1alpha1.MariaDB) (ctrl.Result, error) {
	currentSpec, err := r.getSecret(ctx, mariadb)
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredSpec, err := r.getDesiredSecretSpec(mariadb)
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

func (r *SecretReconciler) getSecret(ctx context.Context, mariadb *v1alpha1.MariaDB) (*corev1.Secret, error) {
	var secret corev1.Secret
	if err := r.GetClient().Get(ctx, client.ObjectKeyFromObject(mariadb), &secret); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &secret, nil
}

func (r *SecretReconciler) getDesiredSecretSpec(mariadb *v1alpha1.MariaDB) (*corev1.Secret, error) {
	rootPassword, err := r.createRandomPassword()
	if err != nil {
		return nil, err
	}
	user, err := r.createRandomUser()
	if err != nil {
		return nil, err
	}
	userPassword, err := r.createRandomPassword()
	if err != nil {
		return nil, err
	}
	database, err := r.createRandomDatabase()
	if err != nil {
		return nil, err
	}
	result := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mariadb.Name,
			Namespace: mariadb.Namespace,
			Labels:    mariadb.GetDesiredLabels(),
		},
		StringData: map[string]string{
			"MARIADB_ROOT_PASSWORD": rootPassword,
			"MARIADB_USER":          user,
			"MARIADB_PASSWORD":      userPassword,
			"MARIADB_DATABASE":      database,
			"MARIADB_AUTO_UPGRADE":  "yes",
		},
	}
	if err := controllerutil.SetControllerReference(mariadb, &result, r.GetClient().Scheme()); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *SecretReconciler) randomStringFromCharset(length int) (string, error) {
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

func (r *SecretReconciler) createRandomPassword() (string, error) {
	return r.randomStringFromCharset(128)
}

// createRandomUser creates a random username for MariaDB.
// See https://mariadb.com/kb/en/create-user/#user-name-component for details on username rules.
func (r *SecretReconciler) createRandomUser() (string, error) {
	return r.randomStringFromCharset(80)
}

// createRandomDatabase creates a random database name for MariaDB.
// See https://mariadb.com/kb/en/identifier-names/ for details on database names.
func (r *SecretReconciler) createRandomDatabase() (string, error) {
	return r.randomStringFromCharset(64)
}
