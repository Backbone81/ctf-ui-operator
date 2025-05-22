package ctfd_test

import (
	"context"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/testutils"
)

var (
	testEnv   *envtest.Environment
	k8sClient client.Client
)

func TestReconciler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ctfd Suite")
}

var _ = BeforeSuite(func() {
	testEnv, k8sClient = testutils.SetupTestEnv()
})

var _ = AfterSuite(func() {
	Expect(testEnv.Stop()).To(Succeed())
})

func DeleteAllInstances(ctx context.Context) {
	var mariadbList v1alpha1.CTFdList
	Expect(k8sClient.List(ctx, &mariadbList)).To(Succeed())

	for _, ctfd := range mariadbList.Items {
		Expect(k8sClient.Delete(ctx, &ctfd)).To(Succeed())
	}
}
