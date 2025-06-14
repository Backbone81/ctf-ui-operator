package ctfd

import (
	"context"
	"slices"

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

	// We need to run update first, so we can deal with challenges which were deleted from the instance and need to
	// be recreated.
	if err := r.updateExistingChallenges(ctx, ctfd, ctfdClient, challengeDescriptionList.Items); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.createMissingChallenges(ctx, ctfd, ctfdClient, challengeDescriptionList.Items); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.deleteRemainingChallenges(ctx, ctfd, ctfdClient); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ChallengeDescriptionReconciler) createMissingChallenges(ctx context.Context, ctfd *v1alpha1.CTFd, ctfdClient *ctfdapi.Client, challengeDescriptions []v1alpha2.ChallengeDescription) error {
	for _, challengeDescription := range challengeDescriptions {
		index := ctfd.Status.GetChallengeDescriptionIndex(challengeDescription)
		if index != -1 {
			continue
		}

		ctrl.LoggerFrom(ctx).Info(
			"Creating challenge",
			"name", challengeDescription.Spec.Title,
		)
		challenge, err := ctfdClient.CreateChallenge(ctx, ctfdapi.Challenge{
			Name:        challengeDescription.Spec.Title,
			Description: challengeDescription.Spec.Description,
			Value:       challengeDescription.Spec.Value,
			Category:    challengeDescription.Spec.Category,
		})
		if err != nil {
			return err
		}
		ctfd.Status.ChallengeDescriptions = append(ctfd.Status.ChallengeDescriptions, v1alpha1.ChallengeDescriptionStatus{
			Id:        challenge.Id,
			Name:      challengeDescription.Name,
			Namespace: challengeDescription.Namespace,
		})
		if err := r.GetClient().Status().Update(ctx, ctfd); err != nil {
			return err
		}
	}
	return nil
}

func (r *ChallengeDescriptionReconciler) updateExistingChallenges(ctx context.Context, ctfd *v1alpha1.CTFd, ctfdClient *ctfdapi.Client, challengeDescriptions []v1alpha2.ChallengeDescription) error {
	for _, challengeDescription := range challengeDescriptions {
		index := ctfd.Status.GetChallengeDescriptionIndex(challengeDescription)
		if index == -1 {
			continue
		}

		challenge, err := ctfdClient.GetChallenge(ctx, ctfd.Status.ChallengeDescriptions[index].Id)
		if err != nil {
			// We might run into an 404 not found error here, if somebody deleted the challenge from the instance. We
			// would need to remove that entry from our status, to have it created afterward. We are unable to detect
			// 404 not found right now.
			return err
		}
		if challenge.Name == challengeDescription.Spec.Title &&
			challenge.Description == challengeDescription.Spec.Description &&
			challenge.Value == challengeDescription.Spec.Value &&
			challenge.Category == challengeDescription.Spec.Category {
			continue
		}

		ctrl.LoggerFrom(ctx).Info(
			"Updating challenge",
			"id", challenge.Id,
			"name", challengeDescription.Spec.Title,
		)
		challenge.Name = challengeDescription.Spec.Title
		challenge.Description = challengeDescription.Spec.Description
		challenge.Value = challengeDescription.Spec.Value
		challenge.Category = challengeDescription.Spec.Category
		if _, err := ctfdClient.UpdateChallenge(ctx, challenge); err != nil {
			return err
		}
	}
	return nil
}

func (r *ChallengeDescriptionReconciler) deleteRemainingChallenges(ctx context.Context, ctfd *v1alpha1.CTFd, ctfdClient *ctfdapi.Client) error {
	challenges, err := ctfdClient.ListChallenges(ctx)
	if err != nil {
		return err
	}
	for _, challenge := range challenges {
		shouldDelete, err := r.shouldDeleteChallenge(ctx, ctfd, challenge)
		if err != nil {
			return err
		}
		if !shouldDelete {
			continue
		}

		ctrl.LoggerFrom(ctx).Info(
			"Deleting challenge",
			"id", challenge.Id,
			"name", challenge.Name,
		)
		if err := ctfdClient.DeleteChallenge(ctx, challenge.Id); err != nil {
			return err
		}
		if err := r.removeBookkeeping(ctx, ctfd, challenge); err != nil {
			return err
		}
	}
	return nil
}

func (r *ChallengeDescriptionReconciler) shouldDeleteChallenge(ctx context.Context, ctfd *v1alpha1.CTFd, challenge ctfdapi.Challenge) (bool, error) {
	index := slices.IndexFunc(ctfd.Status.ChallengeDescriptions, func(status v1alpha1.ChallengeDescriptionStatus) bool {
		return status.Id == challenge.Id
	})
	if index == -1 {
		// Challenges which are not in our bookkeeping need to be deleted.
		return true, nil
	}

	challengeDescription, err := r.getChallengeDescription(ctx, ctfd.Status.ChallengeDescriptions[index].Name, ctfd.Status.ChallengeDescriptions[index].Namespace)
	if err != nil {
		return false, err
	}
	if challengeDescription == nil {
		// Challenges which are in our bookkeeping but have the ChallengeDescription deleted from the cluster need
		// to be deleted.
		return true, nil
	}
	return false, nil
}

func (r *ChallengeDescriptionReconciler) removeBookkeeping(ctx context.Context, ctfd *v1alpha1.CTFd, challenge ctfdapi.Challenge) error {
	index := slices.IndexFunc(ctfd.Status.ChallengeDescriptions, func(status v1alpha1.ChallengeDescriptionStatus) bool {
		return status.Id == challenge.Id
	})
	if index == -1 {
		return nil
	}
	ctfd.Status.ChallengeDescriptions = append(ctfd.Status.ChallengeDescriptions[:index], ctfd.Status.ChallengeDescriptions[index+1:]...)
	if err := r.GetClient().Status().Update(ctx, ctfd); err != nil {
		return err
	}
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

func (r *ChallengeDescriptionReconciler) getChallengeDescription(ctx context.Context, name string, namespace string) (*v1alpha2.ChallengeDescription, error) {
	var challengeDescription v1alpha2.ChallengeDescription
	if err := r.GetClient().Get(
		ctx,
		client.ObjectKey{
			Name:      name,
			Namespace: namespace,
		},
		&challengeDescription,
	); err != nil {
		return nil, client.IgnoreNotFound(err)
	}
	return &challengeDescription, nil
}
