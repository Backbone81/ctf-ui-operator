package ctfd_test

import (
	v1alpha2 "github.com/backbone81/ctf-challenge-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/controller/ctfd"
	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
	"github.com/backbone81/ctf-ui-operator/internal/testutils"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

var _ = Describe("ChallengeDescriptionReconciler", func() {
	var (
		reconciler *utils.Reconciler[*v1alpha1.CTFd]
		ctfdClient *ctfdapi.Client
	)

	BeforeEach(func() {
		reconciler = ctfd.NewReconciler(k8sClient, ctfd.WithChallengeDescriptionReconciler(WithCTFdTestEndpoint(endpointUrl)))
		var err error
		ctfdClient, err = ctfdapi.NewClient(endpointUrl, accessToken)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func(ctx SpecContext) {
		DeleteAllChallengeDescriptions(ctx)
		DeleteAllInstances(ctx)
		challenges, err := ctfdClient.ListChallenges(ctx)
		Expect(err).ToNot(HaveOccurred())
		for _, challenge := range challenges {
			Expect(ctfdClient.DeleteChallenge(ctx, challenge.Id)).Error().ToNot(HaveOccurred())
		}
	})

	It("should successfully create the challenge", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := AddDefaults(v1alpha1.CTFd{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
			Spec: v1alpha1.CTFdSpec{
				ChallengeNamespace: ptr.To(corev1.NamespaceDefault),
			},
		})
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())
		instance.Status.Ready = true
		Expect(k8sClient.Status().Update(ctx, &instance)).To(Succeed())
		Expect(CreateAdminSecret(ctx, &instance, &accessToken)).To(Succeed())
		Expect(CreateChallengeDescription(ctx)).Error().ToNot(HaveOccurred())

		challenges, err := ctfdClient.ListChallenges(ctx)
		Expect(err).ToNot(HaveOccurred())
		challengesBefore := len(challenges)

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		challenges, err = ctfdClient.ListChallenges(ctx)
		Expect(err).ToNot(HaveOccurred())
		challengesAfter := len(challenges)
		Expect(challengesAfter).To(Equal(challengesBefore + 1))
	})

	It("should successfully create the hint", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := AddDefaults(v1alpha1.CTFd{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
			Spec: v1alpha1.CTFdSpec{
				ChallengeNamespace: ptr.To(corev1.NamespaceDefault),
			},
		})
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())
		instance.Status.Ready = true
		Expect(k8sClient.Status().Update(ctx, &instance)).To(Succeed())
		Expect(CreateAdminSecret(ctx, &instance, &accessToken)).To(Succeed())
		challengeDescription, err := CreateChallengeDescription(ctx)
		Expect(err).ToNot(HaveOccurred())
		challengeDescription.Spec.Hints = append(challengeDescription.Spec.Hints, v1alpha2.ChallengeHint{
			Description: "This is a test hint",
		})
		Expect(k8sClient.Update(ctx, challengeDescription)).To(Succeed())

		hints, err := ctfdClient.ListHints(ctx)
		Expect(err).ToNot(HaveOccurred())
		hintsBefore := len(hints)

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		hints, err = ctfdClient.ListHints(ctx)
		Expect(err).ToNot(HaveOccurred())
		hintsAfter := len(hints)
		Expect(hintsAfter).To(Equal(hintsBefore + 1))
	})

	It("should successfully create the flag", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := AddDefaults(v1alpha1.CTFd{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
			Spec: v1alpha1.CTFdSpec{
				ChallengeNamespace: ptr.To(corev1.NamespaceDefault),
			},
		})
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())
		instance.Status.Ready = true
		Expect(k8sClient.Status().Update(ctx, &instance)).To(Succeed())
		Expect(CreateAdminSecret(ctx, &instance, &accessToken)).To(Succeed())
		Expect(CreateChallengeDescription(ctx)).Error().ToNot(HaveOccurred())

		flags, err := ctfdClient.ListFlags(ctx)
		Expect(err).ToNot(HaveOccurred())
		flagsBefore := len(flags)

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		flags, err = ctfdClient.ListFlags(ctx)
		Expect(err).ToNot(HaveOccurred())
		flagsAfter := len(flags)
		Expect(flagsAfter).To(Equal(flagsBefore + 1))
	})

	It("should delete manual created challenges", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := AddDefaults(v1alpha1.CTFd{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
			Spec: v1alpha1.CTFdSpec{
				ChallengeNamespace: ptr.To(corev1.NamespaceDefault),
			},
		})
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())
		Expect(CreateAdminSecret(ctx, &instance, &accessToken)).To(Succeed())
		instance.Status.Ready = true
		Expect(k8sClient.Status().Update(ctx, &instance)).To(Succeed())
		Expect(ctfdClient.CreateChallenge(ctx, ctfdapi.Challenge{
			Name:        "test",
			Description: "test",
		})).Error().ToNot(HaveOccurred())
		challenges, err := ctfdClient.ListChallenges(ctx)
		Expect(err).ToNot(HaveOccurred())
		challengesBefore := len(challenges)

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		challenges, err = ctfdClient.ListChallenges(ctx)
		Expect(err).ToNot(HaveOccurred())
		challengesAfter := len(challenges)
		Expect(challengesAfter).To(Equal(challengesBefore - 1))
	})

	It("should delete manual created hints", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := AddDefaults(v1alpha1.CTFd{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
			Spec: v1alpha1.CTFdSpec{
				ChallengeNamespace: ptr.To(corev1.NamespaceDefault),
			},
		})
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())
		Expect(CreateAdminSecret(ctx, &instance, &accessToken)).To(Succeed())
		instance.Status.Ready = true
		Expect(k8sClient.Status().Update(ctx, &instance)).To(Succeed())
		challengeDescription, err := CreateChallengeDescription(ctx)
		Expect(err).ToNot(HaveOccurred())
		challengeDescription.Spec.Hints = append(challengeDescription.Spec.Hints, v1alpha2.ChallengeHint{
			Description: "This is a test hint",
		})
		Expect(k8sClient.Update(ctx, challengeDescription)).To(Succeed())

		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(&instance), &instance)).To(Succeed())

		Expect(ctfdClient.CreateHint(ctx, ctfdapi.Hint{
			Title:       "Additional hint",
			ChallengeId: instance.Status.ChallengeDescriptions[0].Id,
		})).Error().ToNot(HaveOccurred())

		hints, err := ctfdClient.ListHints(ctx)
		Expect(err).ToNot(HaveOccurred())
		hintsBefore := len(hints)

		By("run the reconciler")
		result, err = reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		hints, err = ctfdClient.ListHints(ctx)
		Expect(err).ToNot(HaveOccurred())
		hintsAfter := len(hints)
		Expect(hintsAfter).To(Equal(hintsBefore - 1))
	})

	It("should delete manual created flags", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := AddDefaults(v1alpha1.CTFd{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
			Spec: v1alpha1.CTFdSpec{
				ChallengeNamespace: ptr.To(corev1.NamespaceDefault),
			},
		})
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())
		Expect(CreateAdminSecret(ctx, &instance, &accessToken)).To(Succeed())
		instance.Status.Ready = true
		Expect(k8sClient.Status().Update(ctx, &instance)).To(Succeed())
		Expect(CreateChallengeDescription(ctx)).Error().ToNot(HaveOccurred())

		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(&instance), &instance)).To(Succeed())

		Expect(ctfdClient.CreateFlag(ctx, ctfdapi.Flag{
			ChallengeId: instance.Status.ChallengeDescriptions[0].Id,
			Content:     "CTF{additional_flag}",
		})).Error().ToNot(HaveOccurred())

		flags, err := ctfdClient.ListFlags(ctx)
		Expect(err).ToNot(HaveOccurred())
		flagsBefore := len(flags)

		By("run the reconciler")
		result, err = reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		flags, err = ctfdClient.ListFlags(ctx)
		Expect(err).ToNot(HaveOccurred())
		flagsAfter := len(flags)
		Expect(flagsAfter).To(Equal(flagsBefore - 1))
	})
})
