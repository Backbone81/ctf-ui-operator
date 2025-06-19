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

	if err := r.reconcileChallenges(ctx, ctfdClient, ctfd); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ChallengeDescriptionReconciler) reconcileChallenges(ctx context.Context, ctfdClient *ctfdapi.Client, ctfd *v1alpha1.CTFd) error {
	ctfdChallenges, err := ctfdClient.ListChallenges(ctx)
	if err != nil {
		return err
	}

	var challengeDescriptionList v1alpha2.ChallengeDescriptionList
	if err := r.GetClient().List(
		ctx,
		&challengeDescriptionList,
		client.InNamespace(r.resolveChallengeNamespace(ctfd)),
	); err != nil {
		return err
	}
	k8sChallenges := challengeDescriptionList.Items

	r.cleanupChallengeStatus(ctfd, k8sChallenges, ctfdChallenges)

	if err := r.updateExistingChallenges(ctx, ctfdClient, ctfdChallenges, ctfd, k8sChallenges); err != nil {
		return err
	}

	if err := r.createMissingChallenges(ctx, ctfdClient, ctfd, k8sChallenges); err != nil {
		return err
	}

	if err := r.deleteObsoleteChallenges(ctx, ctfdClient, ctfdChallenges, ctfd); err != nil {
		return err
	}
	return nil
}

func (r *ChallengeDescriptionReconciler) cleanupChallengeStatus(ctfd *v1alpha1.CTFd, k8sChallenges []v1alpha2.ChallengeDescription, ctfdChallenges []ctfdapi.Challenge) {
	// Remove challenges from the bookkeeping which can not be found in Kubernetes anymore.
	ctfd.Status.ChallengeDescriptions = slices.DeleteFunc(ctfd.Status.ChallengeDescriptions, func(challengeStatus v1alpha1.ChallengeDescriptionStatus) bool {
		return !r.k8sChallengeExists(k8sChallenges, challengeStatus)
	})

	// Remove challenges from the bookkeeping which can not be found in CTFd anymore.
	ctfd.Status.ChallengeDescriptions = slices.DeleteFunc(ctfd.Status.ChallengeDescriptions, func(challengeStatus v1alpha1.ChallengeDescriptionStatus) bool {
		return !r.ctfdChallengeExists(ctfdChallenges, challengeStatus)
	})
}

func (r *ChallengeDescriptionReconciler) getK8sChallengeIndex(k8sChallenges []v1alpha2.ChallengeDescription, challengeStatus v1alpha1.ChallengeDescriptionStatus) int {
	return slices.IndexFunc(k8sChallenges, func(k8sChallenge v1alpha2.ChallengeDescription) bool {
		return k8sChallenge.Name == challengeStatus.Name &&
			k8sChallenge.Namespace == challengeStatus.Namespace
	})
}

func (r *ChallengeDescriptionReconciler) k8sChallengeExists(k8sChallenges []v1alpha2.ChallengeDescription, challengeStatus v1alpha1.ChallengeDescriptionStatus) bool {
	return r.getK8sChallengeIndex(k8sChallenges, challengeStatus) != -1
}

func (r *ChallengeDescriptionReconciler) getCTFdChallengeIndex(ctfdChallenges []ctfdapi.Challenge, challengeStatus v1alpha1.ChallengeDescriptionStatus) int {
	return slices.IndexFunc(ctfdChallenges, func(ctfdChallenge ctfdapi.Challenge) bool {
		return ctfdChallenge.Id == challengeStatus.Id
	})
}

func (r *ChallengeDescriptionReconciler) ctfdChallengeExists(ctfdChallenges []ctfdapi.Challenge, challengeStatus v1alpha1.ChallengeDescriptionStatus) bool {
	return r.getCTFdChallengeIndex(ctfdChallenges, challengeStatus) != -1
}

func (r *ChallengeDescriptionReconciler) updateExistingChallenges(ctx context.Context, ctfdClient *ctfdapi.Client, ctfdChallenges []ctfdapi.Challenge, ctfd *v1alpha1.CTFd, k8sChallenges []v1alpha2.ChallengeDescription) error {
	for challengeStatusIdx, challengeStatus := range ctfd.Status.ChallengeDescriptions {
		ctfdChallengeIndex := r.getCTFdChallengeIndex(ctfdChallenges, challengeStatus)
		ctfdChallenge := ctfdChallenges[ctfdChallengeIndex]

		k8sChallengeIndex := r.getK8sChallengeIndex(k8sChallenges, challengeStatus)
		k8sChallenge := k8sChallenges[k8sChallengeIndex]

		if ctfdChallenge.Name == k8sChallenge.Spec.Title &&
			ctfdChallenge.Description == k8sChallenge.Spec.Description &&
			ctfdChallenge.Value == k8sChallenge.Spec.Value &&
			ctfdChallenge.Category == k8sChallenge.Spec.Category {
			continue
		}

		ctrl.LoggerFrom(ctx).Info(
			"Updating challenge",
			"id", ctfdChallenge.Id,
			"name", k8sChallenge.Spec.Title,
		)
		ctfdChallenge.Name = k8sChallenge.Spec.Title
		ctfdChallenge.Description = k8sChallenge.Spec.Description
		ctfdChallenge.Value = k8sChallenge.Spec.Value
		ctfdChallenge.Category = k8sChallenge.Spec.Category
		if _, err := ctfdClient.UpdateChallenge(ctx, ctfdChallenge); err != nil {
			return err
		}
		if err := r.reconcileHints(ctx, ctfdClient, ctfdChallenge, ctfd, &ctfd.Status.ChallengeDescriptions[challengeStatusIdx], k8sChallenge.Spec.Hints); err != nil {
			return err
		}
		if err := r.reconcileFlag(ctx, ctfdClient, ctfdChallenge, &k8sChallenge); err != nil {
			return err
		}
	}
	return nil
}

func (r *ChallengeDescriptionReconciler) getStatusIndexForK8sChallenge(challengeStatus []v1alpha1.ChallengeDescriptionStatus, k8sChallenge v1alpha2.ChallengeDescription) int {
	return slices.IndexFunc(challengeStatus, func(challengeStatus v1alpha1.ChallengeDescriptionStatus) bool {
		return challengeStatus.Name == k8sChallenge.Name &&
			challengeStatus.Namespace == k8sChallenge.Namespace
	})
}

func (r *ChallengeDescriptionReconciler) statusExistsForK8sChallenge(challengeStatus []v1alpha1.ChallengeDescriptionStatus, k8sChallenge v1alpha2.ChallengeDescription) bool {
	return r.getStatusIndexForK8sChallenge(challengeStatus, k8sChallenge) != -1
}

func (r *ChallengeDescriptionReconciler) createMissingChallenges(ctx context.Context, ctfdClient *ctfdapi.Client, ctfd *v1alpha1.CTFd, k8sChallenges []v1alpha2.ChallengeDescription) error {
	missingChallenges := make([]v1alpha2.ChallengeDescription, len(k8sChallenges))
	copy(missingChallenges, k8sChallenges)
	missingChallenges = slices.DeleteFunc(missingChallenges, func(k8sChallenge v1alpha2.ChallengeDescription) bool {
		return r.statusExistsForK8sChallenge(ctfd.Status.ChallengeDescriptions, k8sChallenge)
	})
	for _, k8sChallenge := range missingChallenges {
		ctrl.LoggerFrom(ctx).Info(
			"Creating challenge",
			"name", k8sChallenge.Spec.Title,
		)
		ctfdChallenge, err := ctfdClient.CreateChallenge(ctx, ctfdapi.Challenge{
			Name:        k8sChallenge.Spec.Title,
			Description: k8sChallenge.Spec.Description,
			Value:       k8sChallenge.Spec.Value,
			Category:    k8sChallenge.Spec.Category,
		})
		if err != nil {
			return err
		}
		ctfd.Status.ChallengeDescriptions = append(ctfd.Status.ChallengeDescriptions, v1alpha1.ChallengeDescriptionStatus{
			Id:        ctfdChallenge.Id,
			Name:      k8sChallenge.Name,
			Namespace: k8sChallenge.Namespace,
		})
		challengeStatusIdx := len(ctfd.Status.ChallengeDescriptions) - 1
		if err := r.GetClient().Status().Update(ctx, ctfd); err != nil {
			return err
		}
		if err := r.reconcileHints(ctx, ctfdClient, ctfdChallenge, ctfd, &ctfd.Status.ChallengeDescriptions[challengeStatusIdx], k8sChallenge.Spec.Hints); err != nil {
			return err
		}
		if err := r.reconcileFlag(ctx, ctfdClient, ctfdChallenge, &k8sChallenge); err != nil {
			return err
		}
	}
	return nil
}

func (r *ChallengeDescriptionReconciler) getStatusIndexForCTFdChallenge(challengeStatus []v1alpha1.ChallengeDescriptionStatus, ctfdChallenge ctfdapi.Challenge) int {
	return slices.IndexFunc(challengeStatus, func(challengeStatus v1alpha1.ChallengeDescriptionStatus) bool {
		return ctfdChallenge.Id == challengeStatus.Id
	})
}

func (r *ChallengeDescriptionReconciler) statusExistsForCTFdChallenge(challengeStatus []v1alpha1.ChallengeDescriptionStatus, ctfdChallenge ctfdapi.Challenge) bool {
	return r.getStatusIndexForCTFdChallenge(challengeStatus, ctfdChallenge) != -1
}

func (r *ChallengeDescriptionReconciler) deleteObsoleteChallenges(ctx context.Context, ctfdClient *ctfdapi.Client, ctfdChallenges []ctfdapi.Challenge, ctfd *v1alpha1.CTFd) error {
	obsoleteChallenges := make([]ctfdapi.Challenge, len(ctfdChallenges))
	copy(obsoleteChallenges, ctfdChallenges)
	obsoleteChallenges = slices.DeleteFunc(obsoleteChallenges, func(ctfdChallenge ctfdapi.Challenge) bool {
		return r.statusExistsForCTFdChallenge(ctfd.Status.ChallengeDescriptions, ctfdChallenge)
	})
	for _, ctfdChallenge := range obsoleteChallenges {
		ctrl.LoggerFrom(ctx).Info(
			"Deleting challenge",
			"id", ctfdChallenge.Id,
			"name", ctfdChallenge.Name,
		)
		if err := ctfdClient.DeleteChallenge(ctx, ctfdChallenge.Id); err != nil {
			return err
		}
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

func (r *ChallengeDescriptionReconciler) reconcileHints(ctx context.Context, ctfdClient *ctfdapi.Client, ctfdChallenge ctfdapi.Challenge, ctfd *v1alpha1.CTFd, challengeStatus *v1alpha1.ChallengeDescriptionStatus, k8sHints []v1alpha2.ChallengeHint) error {
	ctfdHints, err := ctfdClient.ListHintsForChallenge(ctx, ctfdChallenge.Id)
	if err != nil {
		return err
	}

	r.cleanupHintStatus(challengeStatus, k8sHints, ctfdHints)

	if err := r.updateExistingHints(ctx, ctfdClient, ctfdChallenge, challengeStatus, ctfdHints, k8sHints); err != nil {
		return err
	}

	if err := r.createMissingHints(ctx, ctfdClient, ctfdChallenge, ctfd, challengeStatus, k8sHints); err != nil {
		return err
	}

	if err := r.deleteObsoleteHints(ctx, ctfdClient, ctfdChallenge, ctfdHints, challengeStatus); err != nil {
		return err
	}
	return nil
}

func (r *ChallengeDescriptionReconciler) cleanupHintStatus(challengeStatus *v1alpha1.ChallengeDescriptionStatus, k8sHints []v1alpha2.ChallengeHint, ctfdHints []ctfdapi.Hint) {
	// Remove hints from the bookkeeping which can not be found in Kubernetes anymore.
	challengeStatus.Hints = slices.DeleteFunc(challengeStatus.Hints, func(hintStatus v1alpha1.HintStatus) bool {
		return len(k8sHints) <= hintStatus.Index
	})

	// Remove hints from the bookkeeping which can not be found in CTFd anymore.
	challengeStatus.Hints = slices.DeleteFunc(challengeStatus.Hints, func(hintStatus v1alpha1.HintStatus) bool {
		return !r.ctfdHintExists(ctfdHints, hintStatus)
	})
}

func (r *ChallengeDescriptionReconciler) getCTFdHintIndex(ctfdHints []ctfdapi.Hint, hintStatus v1alpha1.HintStatus) int {
	return slices.IndexFunc(ctfdHints, func(ctfdHint ctfdapi.Hint) bool {
		return ctfdHint.Id == hintStatus.Id
	})
}

func (r *ChallengeDescriptionReconciler) ctfdHintExists(ctfdHints []ctfdapi.Hint, hintStatus v1alpha1.HintStatus) bool {
	return r.getCTFdHintIndex(ctfdHints, hintStatus) != -1
}

func (r *ChallengeDescriptionReconciler) updateExistingHints(ctx context.Context, ctfdClient *ctfdapi.Client, ctfdChallenge ctfdapi.Challenge, challengeStatus *v1alpha1.ChallengeDescriptionStatus, ctfdHints []ctfdapi.Hint, k8sHints []v1alpha2.ChallengeHint) error {
	for _, hintStatus := range challengeStatus.Hints {
		ctfdHintIndex := r.getCTFdHintIndex(ctfdHints, hintStatus)
		ctfdHint := ctfdHints[ctfdHintIndex]

		k8sHint := k8sHints[hintStatus.Index]

		if ctfdHint.Title == k8sHint.Description &&
			ctfdHint.Cost == k8sHint.Cost {
			continue
		}
		ctrl.LoggerFrom(ctx).Info(
			"Updating hint",
			"id", ctfdHint.Id,
			"name", ctfdHint.Title,
			"challenge-id", ctfdChallenge.Id,
			"challenge-name", ctfdChallenge.Name,
		)
		ctfdHint.Title = k8sHint.Description
		ctfdHint.Cost = k8sHint.Cost
		if _, err := ctfdClient.UpdateHint(ctx, ctfdHint); err != nil {
			return err
		}
	}
	return nil
}

func (r *ChallengeDescriptionReconciler) createMissingHints(ctx context.Context, ctfdClient *ctfdapi.Client, ctfdChallenge ctfdapi.Challenge, ctfd *v1alpha1.CTFd, challengeStatus *v1alpha1.ChallengeDescriptionStatus, k8sHints []v1alpha2.ChallengeHint) error {
	missingHintIdxs := make([]int, len(k8sHints))
	for i := range missingHintIdxs {
		missingHintIdxs[i] = i
	}
	missingHintIdxs = slices.DeleteFunc(missingHintIdxs, func(k8sHintIdx int) bool {
		index := slices.IndexFunc(challengeStatus.Hints, func(hintStatus v1alpha1.HintStatus) bool {
			return hintStatus.Index == k8sHintIdx
		})
		return index != -1
	})
	for _, missingHintIdx := range missingHintIdxs {
		k8sHint := k8sHints[missingHintIdx]
		ctrl.LoggerFrom(ctx).Info(
			"Creating hint",
			"name", k8sHint.Description,
			"challenge-id", ctfdChallenge.Id,
			"challenge-name", ctfdChallenge.Name,
		)
		ctfdHint, err := ctfdClient.CreateHint(ctx, ctfdapi.Hint{
			ChallengeId: ctfdChallenge.Id,
			Title:       k8sHint.Description,
			Cost:        k8sHint.Cost,
		})
		if err != nil {
			return err
		}
		challengeStatus.Hints = append(challengeStatus.Hints, v1alpha1.HintStatus{
			Id:    ctfdHint.Id,
			Index: missingHintIdx,
		})
		if err := r.GetClient().Status().Update(ctx, ctfd); err != nil {
			return err
		}
	}
	return nil
}

func (r *ChallengeDescriptionReconciler) deleteObsoleteHints(ctx context.Context, ctfdClient *ctfdapi.Client, ctfdChallenge ctfdapi.Challenge, ctfdHints []ctfdapi.Hint, challengeStatus *v1alpha1.ChallengeDescriptionStatus) error {
	obsoleteHints := make([]ctfdapi.Hint, len(ctfdHints))
	copy(obsoleteHints, ctfdHints)
	obsoleteHints = slices.DeleteFunc(obsoleteHints, func(ctfdHint ctfdapi.Hint) bool {
		index := slices.IndexFunc(challengeStatus.Hints, func(k8sHintStatus v1alpha1.HintStatus) bool {
			return k8sHintStatus.Id == ctfdHint.Id
		})
		return index != -1
	})
	for _, ctfdHint := range obsoleteHints {
		ctrl.LoggerFrom(ctx).Info(
			"Deleting hint",
			"id", ctfdHint.Id,
			"name", ctfdHint.Title,
			"challenge-id", ctfdChallenge.Id,
			"challenge-name", ctfdChallenge.Name,
		)
		if err := ctfdClient.DeleteHint(ctx, ctfdHint.Id); err != nil {
			return err
		}
	}
	return nil
}

func (r *ChallengeDescriptionReconciler) reconcileFlag(ctx context.Context, ctfdClient *ctfdapi.Client, ctfdChallenge ctfdapi.Challenge, k8sChallenge *v1alpha2.ChallengeDescription) error {
	flags, err := ctfdClient.ListFlagsForChallenge(ctx, ctfdChallenge.Id)
	if err != nil {
		return err
	}

	// Create missing flag
	if len(flags) < 1 {
		ctrl.LoggerFrom(ctx).Info(
			"Creating flag",
			"challenge-id", ctfdChallenge.Id,
			"challenge-name", ctfdChallenge.Name,
		)
		flag, err := ctfdClient.CreateFlag(ctx, ctfdapi.Flag{
			ChallengeId: ctfdChallenge.Id,
			Content:     k8sChallenge.Spec.Flag,
		})
		if err != nil {
			return err
		}
		flags = append(flags, flag)
	}

	// Delete obsolete flags
	for flagIdx := 1; flagIdx < len(flags); flagIdx++ {
		ctrl.LoggerFrom(ctx).Info(
			"Deleting flag",
			"id", flags[flagIdx].Id,
			"challenge-id", ctfdChallenge.Id,
			"challenge-name", ctfdChallenge.Name,
		)
		if err := ctfdClient.DeleteFlag(ctx, flags[flagIdx].Id); err != nil {
			return err
		}
	}

	// Update existing flag
	if flags[0].Content != k8sChallenge.Spec.Flag {
		flags[0].Content = k8sChallenge.Spec.Flag
		ctrl.LoggerFrom(ctx).Info(
			"Updating flag",
			"id", flags[0].Id,
			"challenge-id", ctfdChallenge.Id,
			"challenge-name", ctfdChallenge.Name,
		)
		if _, err := ctfdClient.UpdateFlag(ctx, flags[0]); err != nil {
			return err
		}
	}
	return nil
}
