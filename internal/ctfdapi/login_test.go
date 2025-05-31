package ctfdapi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
)

var _ = Describe("Login", func() {
	var ctfdClient *ctfdapi.Client

	BeforeEach(func(ctx SpecContext) {
		var err error
		ctfdClient, err = ctfdapi.NewClient(endpointUrl, "")
		Expect(err).ToNot(HaveOccurred())
	})

	It("should successfully login with username", func(ctx SpecContext) {
		Expect(ctfdClient.Login(ctx, ctfdapi.LoginRequest{
			Name:     AdminName,
			Password: AdminPassword,
		})).To(Succeed())
	})

	It("should successfully login with email", func(ctx SpecContext) {
		Expect(ctfdClient.Login(ctx, ctfdapi.LoginRequest{
			Name:     AdminEmail,
			Password: AdminPassword,
		})).To(Succeed())
	})

	It("should fail with wrong password", func(ctx SpecContext) {
		Expect(ctfdClient.Login(ctx, ctfdapi.LoginRequest{
			Name:     AdminName,
			Password: "wrong",
		})).ToNot(Succeed())
	})

	It("should fail with wrong username", func(ctx SpecContext) {
		Expect(ctfdClient.Login(ctx, ctfdapi.LoginRequest{
			Name:     "wrong",
			Password: AdminPassword,
		})).ToNot(Succeed())
	})
})
