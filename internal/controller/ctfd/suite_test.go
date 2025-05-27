package ctfd_test

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/controller/ctfd"
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

func CreateRequiredThirdPartySecrets(ctx context.Context, instance *v1alpha1.CTFd) error {
	mariaDBSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctfd.MariaDBName(instance),
			Namespace: instance.Namespace,
		},
		StringData: map[string]string{
			"MARIADB_USER":     "maria-user",
			"MARIADB_PASSWORD": "maria-password",
			"MARIADB_DATABASE": "maria-database",
		},
	}
	if err := k8sClient.Create(ctx, &mariaDBSecret); err != nil {
		return err
	}

	minioSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ctfd.MinioName(instance),
			Namespace: instance.Namespace,
		},
		StringData: map[string]string{
			"MINIO_ROOT_USER":     "minio-user",
			"MINIO_ROOT_PASSWORD": "minio-password",
		},
	}
	if err := k8sClient.Create(ctx, &minioSecret); err != nil {
		return err
	}
	return nil
}

func SetThirdPartyCRsReady(ctx context.Context, instance *v1alpha1.CTFd, ready bool) error {
	if err := SetRedisReady(ctx, instance, ready); err != nil {
		return err
	}
	if err := SetMariaDBReady(ctx, instance, ready); err != nil {
		return err
	}
	if err := SetMinioReady(ctx, instance, ready); err != nil {
		return err
	}
	return nil
}

func SetRedisReady(ctx context.Context, instance *v1alpha1.CTFd, ready bool) error {
	var redis v1alpha1.Redis
	redis.Name = ctfd.RedisName(instance)
	redis.Namespace = instance.Namespace
	if err := k8sClient.Create(ctx, &redis); err != nil {
		return err
	}
	redis.Status.Ready = ready
	return k8sClient.Status().Update(ctx, &redis)
}

func SetMariaDBReady(ctx context.Context, instance *v1alpha1.CTFd, ready bool) error {
	var mariaDB v1alpha1.MariaDB
	mariaDB.Name = ctfd.MariaDBName(instance)
	mariaDB.Namespace = instance.Namespace
	if err := k8sClient.Create(ctx, &mariaDB); err != nil {
		return err
	}
	mariaDB.Status.Ready = ready
	return k8sClient.Status().Update(ctx, &mariaDB)
}

func SetMinioReady(ctx context.Context, instance *v1alpha1.CTFd, ready bool) error {
	var minio v1alpha1.Minio
	minio.Name = ctfd.MinioName(instance)
	minio.Namespace = instance.Namespace
	if err := k8sClient.Create(ctx, &minio); err != nil {
		return err
	}
	minio.Status.Ready = ready
	return k8sClient.Status().Update(ctx, &minio)
}
