package ctfdapi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
)

var _ = Describe("Flags", func() {
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

	It("should create a new flag", func(ctx SpecContext) {
		beforeFlags, err := ctfdClient.ListFlags(ctx, challenge.Id)
		Expect(err).ToNot(HaveOccurred())

		Expect(ctfdClient.CreateFlag(ctx, ctfdapi.Flag{
			ChallengeId: challenge.Id,
			Content:     "CTF{test_flag}",
		})).Error().ToNot(HaveOccurred())

		afterFlags, err := ctfdClient.ListFlags(ctx, challenge.Id)
		Expect(err).ToNot(HaveOccurred())
		Expect(afterFlags).To(HaveLen(len(beforeFlags) + 1))
	})

	It("should get an existing flag", func(ctx SpecContext) {
		flag, err := ctfdClient.CreateFlag(ctx, ctfdapi.Flag{
			ChallengeId: challenge.Id,
			Content:     "CTF{test_flag}",
		})
		Expect(err).ToNot(HaveOccurred())

		flagGet, err := ctfdClient.GetFlag(ctx, flag.Id)
		Expect(err).ToNot(HaveOccurred())

		Expect(flagGet).To(Equal(flagGet))
	})

	It("should update a flag", func(ctx SpecContext) {
		flag, err := ctfdClient.CreateFlag(ctx, ctfdapi.Flag{
			ChallengeId: challenge.Id,
			Content:     "CTF{test_flag}",
		})
		Expect(err).ToNot(HaveOccurred())

		modifiedContent := "CTF{test_flag_modified}"
		flag.Content = modifiedContent
		updatedFlag, err := ctfdClient.UpdateFlag(ctx, flag)
		Expect(err).ToNot(HaveOccurred())

		Expect(updatedFlag.Content).To(Equal(modifiedContent))
	})

	It("should delete a flag", func(ctx SpecContext) {
		flag, err := ctfdClient.CreateFlag(ctx, ctfdapi.Flag{
			ChallengeId: challenge.Id,
			Content:     "CTF{test_flag}",
		})
		Expect(err).ToNot(HaveOccurred())

		beforeFlags, err := ctfdClient.ListFlags(ctx, challenge.Id)
		Expect(err).ToNot(HaveOccurred())

		Expect(ctfdClient.DeleteFlag(ctx, flag.Id)).To(Succeed())

		afterFlags, err := ctfdClient.ListFlags(ctx, challenge.Id)
		Expect(err).ToNot(HaveOccurred())
		Expect(afterFlags).To(HaveLen(len(beforeFlags) - 1))
	})
})
