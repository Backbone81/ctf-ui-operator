package ctfdapi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
)

var _ = Describe("Token", func() {
	var ctfdClient *ctfdapi.Client

	BeforeEach(func(ctx SpecContext) {
		var err error
		ctfdClient, err = ctfdapi.NewClient(endpointUrl, accessToken)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create a token with login", func(ctx SpecContext) {
		ctfdClient, err := ctfdapi.NewClient(endpointUrl, "")
		Expect(err).ToNot(HaveOccurred())

		Expect(ctfdClient.Login(ctx, ctfdapi.LoginRequest{
			Name:     AdminName,
			Password: AdminPassword,
		})).To(Succeed())

		createTokenResponse, err := ctfdClient.CreateToken(ctx, ctfdapi.CreateTokenRequest{
			Description: "test",
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(createTokenResponse.Data.Value).ToNot(BeZero())
	})

	It("should create a token with access token", func(ctx SpecContext) {
		createTokenResponse, err := ctfdClient.CreateToken(ctx, ctfdapi.CreateTokenRequest{
			Description: "test",
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(createTokenResponse.Data.Value).ToNot(BeZero())
	})

	It("should correctly list access tokens", func(ctx SpecContext) {
		beforeList, err := ctfdClient.ListTokens(ctx)
		Expect(err).ToNot(HaveOccurred())

		Expect(ctfdClient.CreateToken(ctx, ctfdapi.CreateTokenRequest{
			Description: "test",
		})).Error().ToNot(HaveOccurred())

		afterList, err := ctfdClient.ListTokens(ctx)
		Expect(err).ToNot(HaveOccurred())

		Expect(afterList.Data).To(HaveLen(len(beforeList.Data) + 1))
	})

	It("should correctly get access tokens", func(ctx SpecContext) {
		createTokenRequest, err := ctfdClient.CreateToken(ctx, ctfdapi.CreateTokenRequest{
			Description: "test",
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(ctfdClient.GetToken(ctx, createTokenRequest.Data.Id)).Error().ToNot(HaveOccurred())
	})

	It("should correctly delete access token", func(ctx SpecContext) {
		createTokenRequest, err := ctfdClient.CreateToken(ctx, ctfdapi.CreateTokenRequest{
			Description: "test",
		})
		Expect(err).ToNot(HaveOccurred())

		beforeList, err := ctfdClient.ListTokens(ctx)
		Expect(err).ToNot(HaveOccurred())

		Expect(ctfdClient.DeleteToken(ctx, createTokenRequest.Data.Id)).Error().ToNot(HaveOccurred())

		afterList, err := ctfdClient.ListTokens(ctx)
		Expect(err).ToNot(HaveOccurred())

		Expect(afterList.Data).To(HaveLen(len(beforeList.Data) - 1))

		Expect(ctfdClient.GetToken(ctx, createTokenRequest.Data.Id)).Error().To(HaveOccurred())
	})
})
