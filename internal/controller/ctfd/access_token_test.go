package ctfd_test

import (
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

var _ = Describe("AccessTokenReconciler", func() {
	var reconciler *utils.Reconciler[*v1alpha1.CTFd]

	BeforeEach(func() {
		reconciler = ctfd.NewReconciler(k8sClient, ctfd.WithAccessTokenReconciler(WithCTFdTestEndpoint(endpointUrl)))
	})

	AfterEach(func(ctx SpecContext) {
		DeleteAllInstances(ctx)
	})

	It("should successfully create the access token", func(ctx SpecContext) {
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

		adminDetails, err := ctfd.GetAdminDetails(ctx, k8sClient, &instance)
		Expect(err).ToNot(HaveOccurred())
		Expect(adminDetails.AccessToken).To(BeZero())

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		adminDetails, err = ctfd.GetAdminDetails(ctx, k8sClient, &instance)
		Expect(err).ToNot(HaveOccurred())
		Expect(adminDetails.AccessToken).ToNot(BeZero())

		ctfdClient, err := ctfdapi.NewClient(endpointUrl, adminDetails.AccessToken)
		Expect(err).ToNot(HaveOccurred())
		Expect(ctfdClient.ListTokens(ctx)).Error().To(Succeed())
	})
})
