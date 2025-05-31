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

var _ = BeforeSuite(func() {
})

var _ = AfterSuite(func() {
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
