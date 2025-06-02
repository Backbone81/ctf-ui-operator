package ctfd

import (
	"context"

	v1alpha2 "github.com/backbone81/ctf-challenge-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// +kubebuilder:rbac:groups=core.ctf.backbone81,resources=challengedescriptions,verbs=get;list;watch

// ChallengeDescriptionReconciler is responsible for reconciling ChallengeDescription resources into the instance.
type ChallengeDescriptionReconciler struct {
	utils.DefaultSubReconciler
	ctfdEndpoint CTFdEndpointStrategy
}

func NewChallengeDescriptionReconciler(client client.Client, options ...SubReconcilerOption) *ChallengeDescriptionReconciler {
	result := &ChallengeDescriptionReconciler{
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

func (r *ChallengeDescriptionReconciler) Reconcile(ctx context.Context, ctfd *v1alpha1.CTFd) (ctrl.Result, error) {
	if ctfd.Spec.ChallengeNamespace == nil {
		ctrl.LoggerFrom(ctx).V(1).Info("No challenge namespace provided, skipping ChallengeDescriptionReconciler.")
		return ctrl.Result{}, nil
	}
	if !ctfd.Status.Ready {
		// The CTFd instance is not ready. We try again later when the instance is up and running. The next reconcile
		// will be triggered when the status changes.
		ctrl.LoggerFrom(ctx).V(1).Info("CTFd is not ready, skipping ChallengeDescriptionReconciler.")
		return ctrl.Result{}, nil
	}

	adminDetails, err := GetAdminDetails(ctx, r.GetClient(), ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	endpoint, err := r.ctfdEndpoint.GetEndpoint(ctx, ctfd)
	if err != nil {
		return ctrl.Result{}, err
	}

	ctfdClient, err := ctfdapi.NewClient(endpoint, adminDetails.AccessToken)
	if err != nil {
		return ctrl.Result{}, err
	}

	var challengeDescriptionList v1alpha2.ChallengeDescriptionList
	if err := r.GetClient().List(
		ctx,
		&challengeDescriptionList,
		client.InNamespace(r.resolveChallengeNamespace(ctfd)),
	); err != nil {
		return ctrl.Result{}, err
	}
	for _, challengeDescription := range challengeDescriptionList.Items {
		if err := r.reconcileChallengeDescription(ctx, ctfdClient, &challengeDescription); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *ChallengeDescriptionReconciler) reconcileChallengeDescription(ctx context.Context, ctfdClient *ctfdapi.Client, challengeDescription *v1alpha2.ChallengeDescription) error {
	challenge, err := ctfdClient.CreateChallenge(ctx, ctfdapi.Challenge{
		Name:        challengeDescription.Spec.Title,
		Description: challengeDescription.Spec.Description,
		Value:       challengeDescription.Spec.Value,
		Category:    challengeDescription.Spec.Category,
	})
	if err != nil {
		return err
	}
	ctrl.LoggerFrom(ctx).Info(
		"Created challenge",
		"id", challenge.Id,
		"name", challenge.Name,
	)
	return nil
}

func (r *ChallengeDescriptionReconciler) resolveChallengeNamespace(ctfd *v1alpha1.CTFd) string {
	if *ctfd.Spec.ChallengeNamespace == "" {
		return ctfd.Namespace
	}
	return *ctfd.Spec.ChallengeNamespace
}

func (r *ChallengeDescriptionReconciler) SetCTFdEndpoint(endpoint CTFdEndpointStrategy) {
	r.ctfdEndpoint = endpoint
}
