package ctfdapi_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/internal/controller/ctfd"
	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
)

const (
	AdminName     = "admin"
	AdminEmail    = "admin@ctfd.internal"
	AdminPassword = "admin123"
)

func TestReconciler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CTFd API Suite")
}

var (
	container   testcontainers.Container
	endpointUrl string
	accessToken string
)

var _ = BeforeSuite(func(ctx SpecContext) {
	var err error
	container, err = NewTestContainer(ctx)
	Expect(err).ToNot(HaveOccurred())

	endpoint, err := container.Endpoint(ctx, "")
	Expect(err).ToNot(HaveOccurred())
	endpointUrl = "http://" + endpoint

	ctfdClient, err := ctfdapi.NewClient(endpointUrl, "")
	Expect(err).ToNot(HaveOccurred())

	Expect(ctfdClient.Setup(ctx, GetDefaultSetupRequest())).To(Succeed())

	Expect(ctfdClient.Login(ctx, ctfdapi.LoginRequest{
		Name:     AdminName,
		Password: AdminPassword,
	})).To(Succeed())
	createTokenResponse, err := ctfdClient.CreateToken(ctx, ctfdapi.CreateTokenRequest{
		Description: "test",
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(createTokenResponse.Data.Value).ToNot(BeZero())
	accessToken = createTokenResponse.Data.Value
})

var _ = AfterSuite(func(ctx SpecContext) {
	Expect(container.Terminate(ctx)).To(Succeed())
})

func NewTestContainer(ctx context.Context) (testcontainers.Container, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: ctfd.Image,
			ExposedPorts: []string{
				"8000",
			},
			WaitingFor: wait.ForLog("Listening at: http://0.0.0.0:8000"),
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}
	return container, nil
}

func GetDefaultSetupRequest() ctfdapi.SetupRequest {
	return ctfdapi.SetupRequest{
		CTFName:                "Test CTF",
		CTFDescription:         "This is a test CTF.",
		UserMode:               ctfdapi.UserModeTeams,
		ChallengeVisibility:    ctfdapi.ChallengeVisibilityPrivate,
		AccountVisibility:      ctfdapi.AccountVisibilityPrivate,
		ScoreVisibility:        ctfdapi.ScoreVisibilityPrivate,
		RegistrationVisibility: ctfdapi.RegistrationVisibilityPrivate,
		VerifyEmails:           true,
		TeamSize:               nil,
		Name:                   AdminName,
		Email:                  AdminEmail,
		Password:               AdminPassword,
		CTFTheme:               ctfdapi.CTFThemeCoreBeta,
		ThemeColor:             nil,
		Start:                  nil,
		End:                    nil,
	}
}
