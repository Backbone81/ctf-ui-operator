package minio_test

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/controller/minio"
	"github.com/backbone81/ctf-ui-operator/internal/testutils"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

var _ = Describe("PersistentVolumeClaimReconciler", func() {
	var reconciler *utils.Reconciler[*v1alpha1.Minio]

	BeforeEach(func() {
		reconciler = minio.NewReconciler(k8sClient, minio.WithPersistentVolumeClaimReconciler())
	})

	AfterEach(func(ctx SpecContext) {
		DeleteAllInstances(ctx)
	})

	It("should successfully create the persistent volume claim", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := v1alpha1.Minio{
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
		var persistentVolumeClaim corev1.PersistentVolumeClaim
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(&instance), &persistentVolumeClaim)).To(Succeed())
		Expect(persistentVolumeClaim.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(resource.MustParse("128Mi")))
	})

	It("should use the persistent volume claim provided by the minio resource", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := v1alpha1.Minio{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
			Spec: v1alpha1.MinioSpec{
				PersistentVolumeClaim: ptr.To(corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("1Gi"),
						},
					},
				}),
			},
		}
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		var persistentVolumeClaim corev1.PersistentVolumeClaim
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(&instance), &persistentVolumeClaim)).To(Succeed())
		Expect(persistentVolumeClaim.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(resource.MustParse("1Gi")))
	})
})
