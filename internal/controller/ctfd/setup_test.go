package ctfd_test

import (
	"github.com/testcontainers/testcontainers-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/controller/ctfd"
	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
	"github.com/backbone81/ctf-ui-operator/internal/testutils"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

var _ = Describe("SetupReconciler", func() {
	var (
		container   testcontainers.Container
		endpointUrl string

		reconciler *utils.Reconciler[*v1alpha1.CTFd]
	)

	BeforeEach(func(ctx SpecContext) {
		var err error
		container, err = testutils.NewCTFdTestContainer(ctx)
		Expect(err).ToNot(HaveOccurred())

		endpoint, err := container.Endpoint(ctx, "")
		Expect(err).ToNot(HaveOccurred())
		endpointUrl = "http://" + endpoint

		reconciler = ctfd.NewReconciler(k8sClient, ctfd.WithSetupReconciler(WithCTFdTestEndpoint(endpointUrl)))
	})

	AfterEach(func(ctx SpecContext) {
		Expect(container.Terminate(ctx)).To(Succeed())
		DeleteAllInstances(ctx)
	})

	It("should successfully setup instance", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := AddDefaults(v1alpha1.CTFd{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
		})
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())
		instance.Status.Ready = true
		Expect(k8sClient.Status().Update(ctx, &instance)).To(Succeed())

		Expect(CreateAdminSecret(ctx, &instance)).To(Succeed())

		ctfdClient, err := ctfdapi.NewClient(endpointUrl, "")
		Expect(err).ToNot(HaveOccurred())
		Expect(ctfdClient.SetupRequired(ctx)).To(BeTrue())

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		Expect(ctfdClient.SetupRequired(ctx)).To(BeFalse())
	})
})
