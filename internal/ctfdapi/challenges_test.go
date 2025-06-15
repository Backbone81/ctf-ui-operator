package ctfdapi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
)

var _ = Describe("Challenges", func() {
	var ctfdClient *ctfdapi.Client

	BeforeEach(func(ctx SpecContext) {
		var err error
		ctfdClient, err = ctfdapi.NewClient(endpointUrl, accessToken)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create a new challenge", func(ctx SpecContext) {
		beforeChallenges, err := ctfdClient.ListChallenges(ctx)
		Expect(err).ToNot(HaveOccurred())

		Expect(ctfdClient.CreateChallenge(ctx, ctfdapi.Challenge{
			Name: "Test Challenge",
		})).Error().ToNot(HaveOccurred())

		afterChallenges, err := ctfdClient.ListChallenges(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(afterChallenges).To(HaveLen(len(beforeChallenges) + 1))
	})

	It("should get an existing challenge", func(ctx SpecContext) {
		challenge, err := ctfdClient.CreateChallenge(ctx, ctfdapi.Challenge{
			Name: "Test Challenge",
		})
		Expect(err).ToNot(HaveOccurred())

		challengeGet, err := ctfdClient.GetChallenge(ctx, challenge.Id)
		Expect(err).ToNot(HaveOccurred())

		Expect(challengeGet).To(Equal(challenge))
	})

	It("should update a challenge", func(ctx SpecContext) {
		challenge, err := ctfdClient.CreateChallenge(ctx, ctfdapi.Challenge{
			Name: "Test Challenge",
		})
		Expect(err).ToNot(HaveOccurred())

		modifiedName := "Modified Name"
		challenge.Name = modifiedName
		updatedChallenge, err := ctfdClient.UpdateChallenge(ctx, challenge)
		Expect(err).ToNot(HaveOccurred())

		Expect(updatedChallenge.Name).To(Equal(modifiedName))
	})

	It("should delete a challenge", func(ctx SpecContext) {
		challenge, err := ctfdClient.CreateChallenge(ctx, ctfdapi.Challenge{
			Name: "Test Challenge",
		})
		Expect(err).ToNot(HaveOccurred())

		beforeChallenges, err := ctfdClient.ListChallenges(ctx)
		Expect(err).ToNot(HaveOccurred())

		Expect(ctfdClient.DeleteChallenge(ctx, challenge.Id)).To(Succeed())

		afterChallenges, err := ctfdClient.ListChallenges(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(afterChallenges).To(HaveLen(len(beforeChallenges) - 1))
	})
})
