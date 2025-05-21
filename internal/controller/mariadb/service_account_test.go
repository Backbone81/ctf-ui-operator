package mariadb_test

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/controller/mariadb"
	"github.com/backbone81/ctf-ui-operator/internal/testutils"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

var _ = Describe("ServiceAccountReconciler", func() {
	var reconciler *utils.Reconciler[*v1alpha1.MariaDB]

	BeforeEach(func() {
		reconciler = mariadb.NewReconciler(k8sClient, mariadb.WithServiceAccountReconciler())
	})

	AfterEach(func(ctx SpecContext) {
		DeleteAllInstances(ctx)
	})

	It("should successfully create the service account", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := v1alpha1.MariaDB{
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
		var serviceAccount corev1.ServiceAccount
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(&instance), &serviceAccount)).To(Succeed())
		Expect(serviceAccount.AutomountServiceAccountToken).To(HaveValue(BeFalse()))
	})
})
