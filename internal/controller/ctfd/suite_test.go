package ctfd_test

import (
	"context"
	"testing"

	v1alpha2 "github.com/backbone81/ctf-challenge-operator/api/v1alpha1"
	"github.com/testcontainers/testcontainers-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/controller/ctfd"
	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
	"github.com/backbone81/ctf-ui-operator/internal/testutils"
)

const (
	AdminName     = "admin"
	AdminEmail    = "admin@ctfd.internal"
	AdminPassword = "admin123"
)

var (
	testEnv   *envtest.Environment
	k8sClient client.Client

	container   testcontainers.Container
	endpointUrl string
	accessToken string
)

func TestReconciler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CTFd Suite")
}

var _ = BeforeSuite(func(ctx SpecContext) {
	testEnv, k8sClient = testutils.SetupTestEnv()

	var err error
	container, err = testutils.NewCTFdTestContainer(ctx)
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
	Expect(testEnv.Stop()).To(Succeed())
	Expect(container.Terminate(ctx)).To(Succeed())
})

func DeleteAllInstances(ctx context.Context) {
	var mariadbList v1alpha1.CTFdList
	Expect(k8sClient.List(ctx, &mariadbList)).To(Succeed())

	for _, ctfd := range mariadbList.Items {
		Expect(k8sClient.Delete(ctx, &ctfd)).To(Succeed())
	}
}

func DeleteAllChallengeDescriptions(ctx context.Context) {
	var challengeDescriptionList v1alpha2.ChallengeDescriptionList
	Expect(k8sClient.List(ctx, &challengeDescriptionList)).To(Succeed())

	for _, challengeDescription := range challengeDescriptionList.Items {
		Expect(k8sClient.Delete(ctx, &challengeDescription)).To(Succeed())
	}
}

func CreateRequiredThirdPartySecrets(ctx context.Context, instance *v1alpha1.CTFd) error {
	mariaDBSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctfd.MariaDBName(instance),
			Namespace: instance.Namespace,
		},
		StringData: map[string]string{
			"MARIADB_USER":     "maria-user",
			"MARIADB_PASSWORD": "maria-password",
			"MARIADB_DATABASE": "maria-database",
		},
	}
	if err := k8sClient.Create(ctx, &mariaDBSecret); err != nil {
		return err
	}

	minioSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctfd.MinioName(instance),
			Namespace: instance.Namespace,
		},
		StringData: map[string]string{
			"MINIO_ROOT_USER":     testutils.MinioUser,
			"MINIO_ROOT_PASSWORD": testutils.MinioPassword,
		},
	}
	if err := k8sClient.Create(ctx, &minioSecret); err != nil {
		return err
	}
	return nil
}

func CreateAdminSecret(ctx context.Context, instance *v1alpha1.CTFd, accessToken *string) error {
	adminSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctfd.AdminSecretName(instance),
			Namespace: instance.Namespace,
		},
		Data: map[string][]byte{
			"name":     []byte(AdminName),
			"email":    []byte(AdminEmail),
			"password": []byte(AdminPassword),
		},
	}
	if accessToken != nil {
		adminSecret.Data["token"] = []byte(*accessToken)
	}
	if err := k8sClient.Create(ctx, &adminSecret); err != nil {
		return err
	}
	return nil
}

func CreateChallengeDescription(ctx context.Context) (*v1alpha2.ChallengeDescription, error) {
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}
	configMapRaw, err := ToRaw(&configMap)
	if err != nil {
		return nil, err
	}

	challengeDescription := v1alpha2.ChallengeDescription{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-",
			Namespace:    corev1.NamespaceDefault,
		},
		Spec: v1alpha2.ChallengeDescriptionSpec{
			Title:       "Test Challenge",
			Description: "This is a test challenge",
			Manifests: []runtime.RawExtension{
				{
					Raw: configMapRaw,
				},
			},
		},
	}
	if err := k8sClient.Create(ctx, &challengeDescription); err != nil {
		return nil, err
	}
	return &challengeDescription, nil
}

// ToRaw converts a Kubernetes object into its JSON representation.
func ToRaw(obj client.Object) ([]byte, error) {
	codecFactory := serializer.NewCodecFactory(clientgoscheme.Scheme)
	encoder := codecFactory.LegacyCodec(getGroupVersionKind(obj).GroupVersion())
	return runtime.Encode(encoder, obj)
}

func getGroupVersionKind(obj client.Object) schema.GroupVersionKind {
	if !obj.GetObjectKind().GroupVersionKind().Empty() {
		return obj.GetObjectKind().GroupVersionKind()
	}
	gvks, _, err := clientgoscheme.Scheme.ObjectKinds(obj)
	Expect(err).ToNot(HaveOccurred())
	Expect(gvks).To(HaveLen(1))
	return gvks[0]
}

func SetThirdPartyCRsReady(ctx context.Context, instance *v1alpha1.CTFd, ready bool) error {
	if err := SetRedisReady(ctx, instance, ready); err != nil {
		return err
	}
	if err := SetMariaDBReady(ctx, instance, ready); err != nil {
		return err
	}
	if err := SetMinioReady(ctx, instance, ready); err != nil {
		return err
	}
	return nil
}

func SetRedisReady(ctx context.Context, instance *v1alpha1.CTFd, ready bool) error {
	var redis v1alpha1.Redis
	redis.Name = ctfd.RedisName(instance)
	redis.Namespace = instance.Namespace
	if err := k8sClient.Create(ctx, &redis); err != nil {
		return err
	}
	redis.Status.Ready = ready
	return k8sClient.Status().Update(ctx, &redis)
}

func SetMariaDBReady(ctx context.Context, instance *v1alpha1.CTFd, ready bool) error {
	var mariaDB v1alpha1.MariaDB
	mariaDB.Name = ctfd.MariaDBName(instance)
	mariaDB.Namespace = instance.Namespace
	if err := k8sClient.Create(ctx, &mariaDB); err != nil {
		return err
	}
	mariaDB.Status.Ready = ready
	return k8sClient.Status().Update(ctx, &mariaDB)
}

func SetMinioReady(ctx context.Context, instance *v1alpha1.CTFd, ready bool) error {
	var minio v1alpha1.Minio
	minio.Name = ctfd.MinioName(instance)
	minio.Namespace = instance.Namespace
	if err := k8sClient.Create(ctx, &minio); err != nil {
		return err
	}
	minio.Status.Ready = ready
	return k8sClient.Status().Update(ctx, &minio)
}

// AddDefaults sets default values to spec fields if they are not set. This is necessary for tests, as envtest does not
// set default values automatically as specified in the CRD. Full-blown Kubernetes API servers are filling in the
// defaults correctly.
func AddDefaults(ctfd v1alpha1.CTFd) v1alpha1.CTFd {
	addDefaultString(&ctfd.Spec.Title, "Demo CTF")
	addDefaultString(&ctfd.Spec.Description, "This is a demo CTF.")
	addDefaultString(&ctfd.Spec.UserMode, "teams")
	addDefaultString(&ctfd.Spec.ChallengeVisibility, "private")
	addDefaultString(&ctfd.Spec.AccountVisibility, "private")
	addDefaultString(&ctfd.Spec.ScoreVisibility, "private")
	addDefaultString(&ctfd.Spec.RegistrationVisibility, "private")
	addDefaultString(&ctfd.Spec.Theme, "core-beta")
	return ctfd
}

func addDefaultString(text *string, defaultText string) {
	if len(*text) != 0 {
		// We do not overwrite if the value is already set.
		return
	}
	*text = defaultText
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

type TestCTFdEndpointStrategy struct {
	endpointUrl string
}

var _ ctfd.CTFdEndpointStrategy = (*TestCTFdEndpointStrategy)(nil)

func (s *TestCTFdEndpointStrategy) GetEndpoint(ctx context.Context, ctfd *v1alpha1.CTFd) (string, error) {
	return s.endpointUrl, nil
}

func WithCTFdTestEndpoint(endpointUrl string) ctfd.SubReconcilerOption {
	return func(subReconciler any) {
		endpointSetter, ok := subReconciler.(ctfd.CTFdEndpointSetter)
		if !ok {
			panic("this option requires the sub reconciler to implement the CTFdEndpointSetter interface")
		}
		endpointSetter.SetCTFdEndpoint(&TestCTFdEndpointStrategy{
			endpointUrl: endpointUrl,
		})
	}
}

type TestMinioEndpointStrategy struct {
	endpointUrl string
}

var _ ctfd.MinioEndpointStrategy = (*TestMinioEndpointStrategy)(nil)

func (s *TestMinioEndpointStrategy) GetEndpoint(ctx context.Context, ctfd *v1alpha1.CTFd) (string, error) {
	return s.endpointUrl, nil
}

func WithMinioTestEndpoint(endpointUrl string) ctfd.SubReconcilerOption {
	return func(subReconciler any) {
		endpointSetter, ok := subReconciler.(ctfd.MinioEndpointSetter)
		if !ok {
			panic("this option requires the sub reconciler to implement the MinioEndpointSetter interface")
		}
		endpointSetter.SetMinioEndpoint(&TestMinioEndpointStrategy{
			endpointUrl: endpointUrl,
		})
	}
}
