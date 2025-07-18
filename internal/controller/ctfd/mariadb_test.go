package ctfd_test

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/controller/ctfd"
	"github.com/backbone81/ctf-ui-operator/internal/testutils"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

var _ = Describe("MariaDBReconciler", func() {
	var reconciler *utils.Reconciler[*v1alpha1.CTFd]

	BeforeEach(func() {
		reconciler = ctfd.NewReconciler(k8sClient, ctfd.WithMariaDBReconciler())
	})

	AfterEach(func(ctx SpecContext) {
		DeleteAllInstances(ctx)
	})

	It("should successfully create the MariaDB", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := AddDefaults(v1alpha1.CTFd{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
		})
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		var mariadb v1alpha1.MariaDB
		Expect(k8sClient.Get(ctx, client.ObjectKey{
			Namespace: instance.Namespace,
			Name:      ctfd.MariaDBName(&instance),
		}, &mariadb)).To(Succeed())
		Expect(mariadb.Spec.Resources).To(BeZero())
	})

	It("should use the resources provided by the ctfd resource", func(ctx SpecContext) {
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
		instance := AddDefaults(v1alpha1.CTFd{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
			Spec: v1alpha1.CTFdSpec{
				MariaDB: v1alpha1.MariaDBSpec{
					Resources: ptr.To(resources),
				},
			},
		})
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		var mariadb v1alpha1.MariaDB
		Expect(k8sClient.Get(ctx, client.ObjectKey{
			Namespace: instance.Namespace,
			Name:      ctfd.MariaDBName(&instance),
		}, &mariadb)).To(Succeed())
		Expect(mariadb.Spec.Resources).To(HaveValue(Equal(resources)))
	})
})
