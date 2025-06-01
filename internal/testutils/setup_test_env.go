package testutils

import (
	"go.uber.org/zap/zapcore"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck // Dot imports are fine for testutils.
	. "github.com/onsi/gomega"    //nolint:staticcheck // Dot imports are fine for testutils.

	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// SetupTestEnv prepares envtest and a kubernetes client for interacting with the envtest instance. The client is
// uncached.
func SetupTestEnv() (*envtest.Environment, client.Client) {
	Expect(MoveToProjectRoot()).To(Succeed())
	Expect(MakeBinDirAvailable()).To(Succeed())

	logger := zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true), zap.Level(zapcore.Level(-10)))
	ctrllog.SetLogger(logger)

	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{"manifests/ctf-ui-operator-crd.yaml"},
		ErrorIfCRDPathMissing: true,
		BinaryAssetsDirectory: "bin",
	}
	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err := client.New(cfg, client.Options{Scheme: clientgoscheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
	k8sClient = utils.NewLoggingClient(k8sClient)
	return testEnv, k8sClient
}
