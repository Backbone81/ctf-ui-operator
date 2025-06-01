//nolint:goconst // not every string which occurs several times is worth a constant
package ctfdapi_test

import (
	"time"

	"github.com/testcontainers/testcontainers-go"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
	"github.com/backbone81/ctf-ui-operator/internal/testutils"
)

var _ = Describe("Setup", func() {
	var (
		container  testcontainers.Container
		ctfdClient *ctfdapi.Client
	)

	BeforeEach(func(ctx SpecContext) {
		var err error
		container, err = testutils.NewCTFdTestContainer(ctx)
		Expect(err).ToNot(HaveOccurred())

		endpoint, err := container.Endpoint(ctx, "")
		Expect(err).ToNot(HaveOccurred())

		ctfdClient, err = ctfdapi.NewClient("http://"+endpoint, "")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func(ctx SpecContext) {
		Expect(container.Terminate(ctx)).To(Succeed())
	})

	It("should successfully setup CTFd", func(ctx SpecContext) {
		Expect(ctfdClient.SetupRequired(ctx)).To(BeTrue())
		Expect(ctfdClient.Setup(ctx, GetDefaultSetupRequest())).To(Succeed())
		Expect(ctfdClient.SetupRequired(ctx)).To(BeFalse())
	})

	It("should fail setup without required fields", func(ctx SpecContext) {
		Expect(ctfdClient.SetupRequired(ctx)).To(BeTrue())
		Expect(ctfdClient.Setup(ctx, ctfdapi.SetupRequest{})).ToNot(Succeed())
		Expect(ctfdClient.SetupRequired(ctx)).To(BeTrue())
	})

	It("should fail without a name", func(ctx SpecContext) {
		setupRequest := GetDefaultSetupRequest()
		setupRequest.Name = ""
		Expect(ctfdClient.Setup(ctx, setupRequest)).ToNot(Succeed())
	})

	DescribeTable("should accept valid descriptions",
		func(ctx SpecContext, description string) {
			setupRequest := GetDefaultSetupRequest()
			setupRequest.CTFDescription = description
			Expect(ctfdClient.Setup(ctx, setupRequest)).To(Succeed())
		},
		Entry("when description is empty", ""),
		Entry("when description is provided", "This is a description."),
	)

	DescribeTable("should accept all possible user modes",
		func(ctx SpecContext, userMode ctfdapi.UserMode) {
			setupRequest := GetDefaultSetupRequest()
			setupRequest.UserMode = userMode
			Expect(ctfdClient.Setup(ctx, setupRequest)).To(Succeed())
		},
		Entry("when user mode is teams", ctfdapi.UserModeTeams),
		Entry("when user mode is users", ctfdapi.UserModeUsers),
	)

	DescribeTable("should accept all possible challenge visibilities",
		func(ctx SpecContext, challengeVisibility ctfdapi.ChallengeVisibility) {
			setupRequest := GetDefaultSetupRequest()
			setupRequest.ChallengeVisibility = challengeVisibility
			Expect(ctfdClient.Setup(ctx, setupRequest)).To(Succeed())
		},
		Entry("when challenge visibility is public", ctfdapi.ChallengeVisibilityPublic),
		Entry("when challenge visibility is private", ctfdapi.ChallengeVisibilityPrivate),
		Entry("when challenge visibility is admins", ctfdapi.ChallengeVisibilityAdmins),
	)

	It("should fail with an invalid challenge visibility", func(ctx SpecContext) {
		setupRequest := GetDefaultSetupRequest()
		setupRequest.ChallengeVisibility = "invalid"
		Expect(ctfdClient.Setup(ctx, setupRequest)).ToNot(Succeed())
	})

	DescribeTable("should accept all possible account visibilities",
		func(ctx SpecContext, accountVisibility ctfdapi.AccountVisibility) {
			setupRequest := GetDefaultSetupRequest()
			setupRequest.AccountVisibility = accountVisibility
			Expect(ctfdClient.Setup(ctx, setupRequest)).To(Succeed())
		},
		Entry("when account visibility is public", ctfdapi.AccountVisibilityPublic),
		Entry("when account visibility is private", ctfdapi.AccountVisibilityPrivate),
		Entry("when account visibility is admins", ctfdapi.AccountVisibilityAdmins),
	)

	It("should fail with an invalid account visibility", func(ctx SpecContext) {
		setupRequest := GetDefaultSetupRequest()
		setupRequest.AccountVisibility = "invalid"
		Expect(ctfdClient.Setup(ctx, setupRequest)).ToNot(Succeed())
	})

	DescribeTable("should accept all possible score visibilities",
		func(ctx SpecContext, scoreVisibility ctfdapi.ScoreVisibility) {
			setupRequest := GetDefaultSetupRequest()
			setupRequest.ScoreVisibility = scoreVisibility
			Expect(ctfdClient.Setup(ctx, setupRequest)).To(Succeed())
		},
		Entry("when score visibility is public", ctfdapi.ScoreVisibilityPublic),
		Entry("when score visibility is private", ctfdapi.ScoreVisibilityPrivate),
		Entry("when score visibility is hidden", ctfdapi.ScoreVisibilityHidden),
		Entry("when score visibility is admins", ctfdapi.ScoreVisibilityAdmins),
	)

	It("should fail with an invalid score visibility", func(ctx SpecContext) {
		setupRequest := GetDefaultSetupRequest()
		setupRequest.ScoreVisibility = "invalid"
		Expect(ctfdClient.Setup(ctx, setupRequest)).ToNot(Succeed())
	})

	DescribeTable("should accept all possible registration visibilities",
		func(ctx SpecContext, registrationVisibility ctfdapi.RegistrationVisibility) {
			setupRequest := GetDefaultSetupRequest()
			setupRequest.RegistrationVisibility = registrationVisibility
			Expect(ctfdClient.Setup(ctx, setupRequest)).To(Succeed())
		},
		Entry("when registration visibility is public", ctfdapi.RegistrationVisibilityPublic),
		Entry("when registration visibility is private", ctfdapi.RegistrationVisibilityPrivate),
		Entry("when registration visibility is mlc", ctfdapi.RegistrationVisibilityMLC),
	)

	It("should fail with an invalid registration visibility", func(ctx SpecContext) {
		setupRequest := GetDefaultSetupRequest()
		setupRequest.RegistrationVisibility = "invalid"
		Expect(ctfdClient.Setup(ctx, setupRequest)).ToNot(Succeed())
	})

	DescribeTable("should accept all possible email verification settings",
		func(ctx SpecContext, verifyEmails bool) {
			setupRequest := GetDefaultSetupRequest()
			setupRequest.VerifyEmails = verifyEmails
			Expect(ctfdClient.Setup(ctx, setupRequest)).To(Succeed())
		},
		Entry("when email verification is disabled", false),
		Entry("when email verification is enabled", true),
	)

	DescribeTable("should accept valid team sizes",
		func(ctx SpecContext, teamSize *int) {
			setupRequest := GetDefaultSetupRequest()
			setupRequest.TeamSize = teamSize
			Expect(ctfdClient.Setup(ctx, setupRequest)).To(Succeed())
		},
		Entry("when no team size is given", nil),
		Entry("when teams are small", ptr.To(3)),
		Entry("when teams are medium", ptr.To(8)),
		Entry("when teams are large", ptr.To(16)),
	)

	It("should fail without an admin name", func(ctx SpecContext) {
		setupRequest := GetDefaultSetupRequest()
		setupRequest.Name = ""
		Expect(ctfdClient.Setup(ctx, setupRequest)).ToNot(Succeed())
	})

	It("should fail without an admin email", func(ctx SpecContext) {
		setupRequest := GetDefaultSetupRequest()
		setupRequest.Email = ""
		Expect(ctfdClient.Setup(ctx, setupRequest)).ToNot(Succeed())
	})

	It("should fail with an invalid admin email", func(ctx SpecContext) {
		setupRequest := GetDefaultSetupRequest()
		setupRequest.Email = "invalid"
		Expect(ctfdClient.Setup(ctx, setupRequest)).ToNot(Succeed())
	})

	It("should fail without an admin password", func(ctx SpecContext) {
		setupRequest := GetDefaultSetupRequest()
		setupRequest.Password = ""
		Expect(ctfdClient.Setup(ctx, setupRequest)).ToNot(Succeed())
	})

	DescribeTable("should accept all possible themes",
		func(ctx SpecContext, ctfTheme ctfdapi.CTFTheme) {
			setupRequest := GetDefaultSetupRequest()
			setupRequest.CTFTheme = ctfTheme
			Expect(ctfdClient.Setup(ctx, setupRequest)).To(Succeed())
		},
		Entry("when theme is core-beta", ctfdapi.CTFThemeCoreBeta),
		Entry("when theme is core", ctfdapi.CTFThemeCore),
	)

	DescribeTable("should accept valid theme colors",
		func(ctx SpecContext, themeColor *string) {
			setupRequest := GetDefaultSetupRequest()
			setupRequest.ThemeColor = themeColor
			Expect(ctfdClient.Setup(ctx, setupRequest)).To(Succeed())
		},
		Entry("when no theme color is given", nil),
		Entry("when theme color is black", ptr.To("#000000")),
		Entry("when theme color is white", ptr.To("#FFFFFF")),
		Entry("when theme color is red", ptr.To("#FF0000")),
		Entry("when theme color is green", ptr.To("#00FF00")),
		Entry("when theme color is blue", ptr.To("#0000FF")),
	)

	DescribeTable("should accept valid start and end times",
		func(ctx SpecContext, start *time.Time, end *time.Time) {
			setupRequest := GetDefaultSetupRequest()
			setupRequest.Start = start
			setupRequest.End = end
			Expect(ctfdClient.Setup(ctx, setupRequest)).To(Succeed())
		},
		Entry("when no start and end is given", nil, nil),
		Entry(
			"when start and end is in the future",
			ptr.To(time.Now().AddDate(0, 0, 3)),
			ptr.To(time.Now().AddDate(0, 0, 5)),
		),
	)
})
