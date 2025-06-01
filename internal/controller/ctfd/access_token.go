package ctfd

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// AccessTokenReconciler is responsible for creating an access token for the operator.
type AccessTokenReconciler struct {
	utils.DefaultSubReconciler
	ctfdEndpoint CTFdEndpointStrategy
}

func NewAccessTokenReconciler(client client.Client, options ...SubReconcilerOption) *AccessTokenReconciler {
	result := &AccessTokenReconciler{
		DefaultSubReconciler: utils.NewDefaultSubReconciler(client),
	}
	for _, option := range options {
		option(result)
	}

	if result.ctfdEndpoint == nil {
		panic("CTFd endpoint strategy required")
	}
	return result
}

func (r *AccessTokenReconciler) Reconcile(ctx context.Context, ctfd *v1alpha1.CTFd) (ctrl.Result, error) {
	if !ctfd.Status.Ready {
		// The CTFd instance is not ready. We try again later when the instance is up and running. The next reconcile
		// will be triggered when the status changes.
		ctrl.LoggerFrom(ctx).V(1).Info("CTFd is not ready, skipping AccessTokenReconciler.")
		return ctrl.Result{}, nil
	}

	adminDetails, err := GetAdminDetails(ctx, r.GetClient(), ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}
	if len(adminDetails.AccessToken) != 0 {
		// We already have an access token for the admin. No need to create another one.
		ctrl.LoggerFrom(ctx).V(1).Info("Access token already available, skipping AccessTokenReconciler.")
		return ctrl.Result{}, nil
	}

	endpoint, err := r.ctfdEndpoint.GetEndpoint(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	ctfdClient, err := ctfdapi.NewClient(endpoint, "")
	if err != nil {
		return ctrl.Result{}, err
	}

	ctrl.LoggerFrom(ctx).Info("Creating access token")
	if err := ctfdClient.Login(ctx, ctfdapi.LoginRequest{
		Name:     adminDetails.Name,
		Password: adminDetails.Password,
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("logging into CTFd: %w", err)
	}

	// NOTE: We are creating an access token with 6 months expiration. This should be long enough for any CTF event to
	// be prepared and finished. We do not implement any automated token refresh right now. If 6 months is not long
	// enough for you, you need to delete the token from the admin secret before the expiration is reached, which will
	// make this reconiler create a new token with 6 months expiration.
	createTokenResponse, err := ctfdClient.CreateToken(ctx, ctfdapi.CreateTokenRequest{
		Description: ctfd.Name + " (ctf-ui-operator)",
		Expiration:  ctfdapi.NewDateOnly(time.Now().AddDate(0, 6, 0)),
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("creating access token: %w", err)
	}

	if err := r.storeAccessToken(ctx, ctfd, createTokenResponse.Data.Value); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AccessTokenReconciler) storeAccessToken(ctx context.Context, ctfd *v1alpha1.CTFd, accessToken string) error {
	var secret corev1.Secret
	if err := r.GetClient().Get(ctx, client.ObjectKey{
		Name:      AdminSecretName(ctfd),
		Namespace: ctfd.Namespace,
	}, &secret); err != nil {
		return err
	}

	secret.Data["token"] = []byte(accessToken)

	if err := r.GetClient().Update(ctx, &secret); err != nil {
		return err
	}
	return nil
}

func (r *AccessTokenReconciler) SetCTFdEndpoint(endpoint CTFdEndpointStrategy) {
	r.ctfdEndpoint = endpoint
}
