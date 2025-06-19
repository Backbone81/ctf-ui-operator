package ctfdapi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
)

var _ = Describe("Hints", func() {
	var (
		ctfdClient *ctfdapi.Client
		challenge  ctfdapi.Challenge
	)

	BeforeEach(func(ctx SpecContext) {
		var err error
		ctfdClient, err = ctfdapi.NewClient(endpointUrl, accessToken)
		Expect(err).ToNot(HaveOccurred())

		challenge, err = ctfdClient.CreateChallenge(ctx, ctfdapi.Challenge{
			Name: "Test Challenge",
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create a new hint", func(ctx SpecContext) {
		beforeHints, err := ctfdClient.ListHintsForChallenge(ctx, challenge.Id)
		Expect(err).ToNot(HaveOccurred())

		Expect(ctfdClient.CreateHint(ctx, ctfdapi.Hint{
			ChallengeId: challenge.Id,
			Content:     "This is a test hint.",
		})).Error().ToNot(HaveOccurred())

		afterHints, err := ctfdClient.ListHintsForChallenge(ctx, challenge.Id)
		Expect(err).ToNot(HaveOccurred())
		Expect(afterHints).To(HaveLen(len(beforeHints) + 1))
	})

	It("should get an existing hint", func(ctx SpecContext) {
		hint, err := ctfdClient.CreateHint(ctx, ctfdapi.Hint{
			ChallengeId: challenge.Id,
			Content:     "This is a test hint.",
		})
		Expect(err).ToNot(HaveOccurred())

		hintGet, err := ctfdClient.GetHint(ctx, hint.Id)
		Expect(err).ToNot(HaveOccurred())

		Expect(hintGet).To(Equal(hint))
	})

	It("should update a hint", func(ctx SpecContext) {
		hint, err := ctfdClient.CreateHint(ctx, ctfdapi.Hint{
			ChallengeId: challenge.Id,
			Content:     "This is a test hint.",
		})
		Expect(err).ToNot(HaveOccurred())

		modifiedContent := "This is a modified test hint."
		hint.Content = modifiedContent
		updatedHint, err := ctfdClient.UpdateHint(ctx, hint)
		Expect(err).ToNot(HaveOccurred())

		Expect(updatedHint.Content).To(Equal(modifiedContent))
	})

	It("should delete a hint", func(ctx SpecContext) {
		hint, err := ctfdClient.CreateHint(ctx, ctfdapi.Hint{
			ChallengeId: challenge.Id,
			Content:     "This is a test hint.",
		})
		Expect(err).ToNot(HaveOccurred())

		beforeHints, err := ctfdClient.ListHintsForChallenge(ctx, challenge.Id)
		Expect(err).ToNot(HaveOccurred())

		Expect(ctfdClient.DeleteHint(ctx, hint.Id)).To(Succeed())

		afterHints, err := ctfdClient.ListHintsForChallenge(ctx, challenge.Id)
		Expect(err).ToNot(HaveOccurred())
		Expect(afterHints).To(HaveLen(len(beforeHints) - 1))
	})
})
