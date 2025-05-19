package redis_test

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/controller/redis"
	"github.com/backbone81/ctf-ui-operator/internal/testutils"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

var _ = Describe("ServiceReconciler", func() {
	var reconciler *utils.Reconciler[*v1alpha1.Redis]

	BeforeEach(func() {
		reconciler = redis.NewReconciler(k8sClient, redis.WithServiceReconciler())
	})

	AfterEach(func(ctx SpecContext) {
		DeleteAllInstances(ctx)
	})

	It("should successfully create the service", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := v1alpha1.Redis{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
		}
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		var service corev1.Service
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(&instance), &service)).To(Succeed())
		Expect(service.Spec.Ports[0].Name).To(Equal("redis"))
	})
})
