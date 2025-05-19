package redis_test

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/controller/redis"
	"github.com/backbone81/ctf-ui-operator/internal/testutils"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

var _ = Describe("DeploymentReconciler", func() {
	var reconciler *utils.Reconciler[*v1alpha1.Redis]

	BeforeEach(func() {
		reconciler = redis.NewReconciler(k8sClient, redis.WithDeploymentReconciler())
	})

	AfterEach(func(ctx SpecContext) {
		DeleteAllInstances(ctx)
	})

	It("should successfully create the deployment", func(ctx SpecContext) {
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
		var deployment appsv1.Deployment
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(&instance), &deployment)).To(Succeed())
		Expect(deployment.Spec.Template.Spec.Containers[0].Resources).To(BeZero())
	})

	It("should use the resources provided by the redis resource", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		resources := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("3"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			},
		}
		instance := v1alpha1.Redis{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
			Spec: v1alpha1.RedisSpec{
				Resources: ptr.To(resources),
			},
		}
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		var deployment appsv1.Deployment
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(&instance), &deployment)).To(Succeed())
		Expect(deployment.Spec.Template.Spec.Containers[0].Resources).To(Equal(resources))
	})
})
