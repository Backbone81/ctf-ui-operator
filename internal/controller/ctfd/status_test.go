package ctfd_test

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

var _ = Describe("StatusReconciler", func() {
	var reconciler *utils.Reconciler[*v1alpha1.CTFd]

	BeforeEach(func() {
		reconciler = ctfd.NewReconciler(k8sClient, ctfd.WithStatusReconciler())
	})

	AfterEach(func(ctx SpecContext) {
		DeleteAllInstances(ctx)
	})

	It("should set the status to not ready when not all replicas are ready", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := v1alpha1.CTFd{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
		}
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())
		deployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.Name,
				Namespace: instance.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](1),
				Selector: ptr.To(metav1.LabelSelector{
					MatchLabels: instance.GetDesiredLabels(),
				}),
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: instance.GetDesiredLabels(),
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "ctfd",
								Image: "foo:bar",
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, &deployment)).To(Succeed())
		deployment.Status.Replicas = 3
		deployment.Status.ReadyReplicas = 2
		Expect(k8sClient.Status().Update(ctx, &deployment)).To(Succeed())
		Expect(SetThirdPartyCRsReady(ctx, &instance, true)).To(Succeed())

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(&instance), &instance)).To(Succeed())
		Expect(instance.Status.Ready).To(BeFalse())
	})

	It("should set the status to ready when all replicas are ready", func(ctx SpecContext) {
		By("prepare test with all preconditions")
		instance := v1alpha1.CTFd{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
		}
		Expect(k8sClient.Create(ctx, &instance)).To(Succeed())
		deployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.Name,
				Namespace: instance.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](1),
				Selector: ptr.To(metav1.LabelSelector{
					MatchLabels: instance.GetDesiredLabels(),
				}),
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: instance.GetDesiredLabels(),
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "ctfd",
								Image: "foo:bar",
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, &deployment)).To(Succeed())
		deployment.Status.Replicas = 3
		deployment.Status.ReadyReplicas = 3
		Expect(k8sClient.Status().Update(ctx, &deployment)).To(Succeed())
		Expect(SetThirdPartyCRsReady(ctx, &instance, true)).To(Succeed())

		By("run the reconciler")
		result, err := reconciler.Reconcile(ctx, testutils.RequestFromObject(&instance))
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(BeZero())

		By("verify all postconditions")
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(&instance), &instance)).To(Succeed())
		Expect(instance.Status.Ready).To(BeTrue())
	})
})
