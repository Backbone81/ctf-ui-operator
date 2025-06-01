package ctfd_test

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/controller/ctfd"
	"github.com/backbone81/ctf-ui-operator/internal/testutils"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

var _ = Describe("MinioBucketReconciler", func() {
	var (
		container   testcontainers.Container
		endpointUrl string

		reconciler *utils.Reconciler[*v1alpha1.CTFd]
	)

	BeforeEach(func(ctx SpecContext) {
		var err error
		container, err = testutils.NewMinioTestContainer(ctx)
		Expect(err).ToNot(HaveOccurred())

		endpoint, err := container.Endpoint(ctx, "")
		Expect(err).ToNot(HaveOccurred())
		endpointUrl = endpoint

		reconciler = ctfd.NewReconciler(k8sClient, ctfd.WithMinioBucketReconciler(WithMinioTestEndpoint(endpointUrl)))
	})

	AfterEach(func(ctx SpecContext) {
		Expect(container.Terminate(ctx)).To(Succeed())
		DeleteAllInstances(ctx)
	})

	It("should successfully create the Minio bucket", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := AddDefaults(v1alpha1.CTFd{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
		})
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())

		Expect(SetThirdPartyCRsReady(ctx, &instance, true)).To(Succeed())
		Expect(CreateRequiredThirdPartySecrets(ctx, &instance)).To(Succeed())

		minioClient, err := minio.New(endpointUrl, &minio.Options{
			Creds:  credentials.NewStaticV4(testutils.MinioUser, testutils.MinioPassword, ""),
			Secure: false,
		})
		Expect(err).ToNot(HaveOccurred())
		beforeBuckets, err := minioClient.ListBuckets(ctx)
		Expect(err).ToNot(HaveOccurred())

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		afterBuckets, err := minioClient.ListBuckets(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(afterBuckets).To(HaveLen(len(beforeBuckets) + 1))
	})
})
