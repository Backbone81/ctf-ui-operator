package ctfd

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// SetupReconciler is responsible for initializing CTFd instance.
type SetupReconciler struct {
	utils.DefaultSubReconciler
	endpointStrategy EndpointStrategy
}

func NewSetupReconciler(client client.Client) *SetupReconciler {
	var endpointStrategy EndpointStrategy
	if _, err := rest.InClusterConfig(); err != nil {
		endpointStrategy = &OutOfClusterEndpointStrategy{
			servicePortForwarder: utils.NewServicePortForwarder(client),
		}
	} else {
		endpointStrategy = &InClusterEndpointStrategy{}
	}

	return &SetupReconciler{
		DefaultSubReconciler: utils.NewDefaultSubReconciler(client),
		endpointStrategy:     endpointStrategy,
	}
}

func (r *SetupReconciler) Reconcile(ctx context.Context, ctfd *v1alpha1.CTFd) (ctrl.Result, error) {
	if !ctfd.Status.Ready {
		// The CTFd instance is not ready. We try again later when the instance is up and running. The next reconcile
		// will be triggered when the status changes.
		return ctrl.Result{}, nil
	}

	endpoint, err := r.getEndpoint(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	ctfdClient, err := ctfdapi.NewClient(endpoint)
	if err != nil {
		return ctrl.Result{}, err
	}

	setupRequired, err := ctfdClient.SetupRequired(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("checking if setup is required: %w", err)
	}
	if !setupRequired {
		// The setup was already done. No need to do anything here.
		ctrl.LoggerFrom(ctx).V(5).Info("Instance is already setup. Not running setup again.")
		return ctrl.Result{}, nil
	}

	ctrl.LoggerFrom(ctx).Info("Setting up CTFd")
	setupRequest, err := r.getSetupRequest(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}
	if err := ctfdClient.Setup(ctx, setupRequest); err != nil {
		return ctrl.Result{}, err
	}

	// We do not get any good feedback when things go wrong during setup. It is basically an HTTP 200 OK but with a
	// box on the website containing the error message. To be sure that we did in fact successfully set up the instance,
	// we double-check again if the /setup route is still available.
	setupRequired, err = ctfdClient.SetupRequired(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("checking if setup is required: %w", err)
	}
	if setupRequired {
		return ctrl.Result{}, errors.New("setup failed: manually check the setup process and make sure that given field values satisfy form validation")
	}
	return ctrl.Result{}, nil
}

func (r *SetupReconciler) getEndpoint(ctx context.Context, ctfd *v1alpha1.CTFd) (string, error) {
	return r.endpointStrategy.GetEndpoint(ctx, ctfd)
}

func (r *SetupReconciler) getSetupRequest(ctx context.Context, ctfd *v1alpha1.CTFd) (ctfdapi.SetupRequest, error) {
	adminDetails, err := GetAdminDetails(ctx, r.GetClient(), ctfd)
	if err != nil {
		return ctfdapi.SetupRequest{}, err
	}
	result := ctfdapi.SetupRequest{
		CTFName:                ctfd.Spec.Title,
		CTFDescription:         ctfd.Spec.Description,
		UserMode:               ctfdapi.UserMode(ctfd.Spec.UserMode),
		ChallengeVisibility:    ctfdapi.ChallengeVisibility(ctfd.Spec.ChallengeVisibility),
		AccountVisibility:      ctfdapi.AccountVisibility(ctfd.Spec.AccountVisibility),
		ScoreVisibility:        ctfdapi.ScoreVisibility(ctfd.Spec.ScoreVisibility),
		RegistrationVisibility: ctfdapi.RegistrationVisibility(ctfd.Spec.RegistrationVisibility),
		VerifyEmails:           ctfd.Spec.VerifyEmails,
		TeamSize:               ctfd.Spec.TeamSize,
		Name:                   adminDetails.Name,
		Email:                  adminDetails.Email,
		Password:               adminDetails.Password,
		CTFTheme:               ctfdapi.CTFTheme(ctfd.Spec.Theme),
		ThemeColor:             ctfd.Spec.ThemeColor,
	}
	if ctfd.Spec.Start != nil {
		result.Start = &ctfd.Spec.Start.Time
	}
	if ctfd.Spec.End != nil {
		result.End = &ctfd.Spec.End.Time
	}
	return result, nil
}
