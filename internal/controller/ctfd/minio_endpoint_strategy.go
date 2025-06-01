package ctfd

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// MinioEndpointStrategy describes the way to get the endpoint of Minio. As we need to differentiate between running
// in-cluster and running out-of-cluster, both strategies need to implement this interface.
//
//nolint:iface // CTFdEndpoint strategy and MinioEndpointStrategy have identical methods by design.
type MinioEndpointStrategy interface {
	GetEndpoint(ctx context.Context, ctfd *v1alpha1.CTFd) (string, error)
}

// InClusterMinioEndpointStrategy returns an endpoint which is the service name and the port for in-cluster usage.
type InClusterMinioEndpointStrategy struct{}

var _ MinioEndpointStrategy = (*InClusterMinioEndpointStrategy)(nil)

func (s *InClusterMinioEndpointStrategy) GetEndpoint(ctx context.Context, ctfd *v1alpha1.CTFd) (string, error) {
	return fmt.Sprintf("%s.%s:9000", MinioName(ctfd), ctfd.Namespace), nil
}

// OutOfClusterMinioEndpointStrategy port forwards the minio service to the local host and returns an endpoint with
// that forwarded port. The local port is a random free port.
type OutOfClusterMinioEndpointStrategy struct {
	servicePortForwarder *utils.ServicePortForwarder
}

var _ MinioEndpointStrategy = (*OutOfClusterMinioEndpointStrategy)(nil)

func (s *OutOfClusterMinioEndpointStrategy) GetEndpoint(ctx context.Context, ctfd *v1alpha1.CTFd) (string, error) {
	localPort, err := s.servicePortForwarder.PortForward(
		ctx,
		types.NamespacedName{
			Namespace: ctfd.Namespace,
			Name:      MinioName(ctfd),
		},
		intstr.FromString("minio"),
	)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("127.0.0.1:%d", localPort), nil
}

type MinioEndpointSetter interface {
	SetMinioEndpoint(minioEndpointStrategy MinioEndpointStrategy)
}

func WithMinioInClusterEndpoint() SubReconcilerOption {
	return func(subReconciler any) {
		endpointSetter, ok := subReconciler.(MinioEndpointSetter)
		if !ok {
			panic("this option requires the sub reconciler to implement the MinioEndpointSetter interface")
		}
		endpointSetter.SetMinioEndpoint(&InClusterMinioEndpointStrategy{})
	}
}

func WithMinioOutOfClusterEndpoint() SubReconcilerOption {
	return func(subReconciler any) {
		clientGetter, ok := subReconciler.(ClientGetter)
		if !ok {
			panic("this option requires the sub reconciler to implement the ClientGetter interface")
		}
		endpointSetter, ok := subReconciler.(MinioEndpointSetter)
		if !ok {
			panic("this option requires the sub reconciler to implement the MinioEndpointSetter interface")
		}
		endpointSetter.SetMinioEndpoint(&OutOfClusterMinioEndpointStrategy{
			servicePortForwarder: utils.NewServicePortForwarder(clientGetter.GetClient()),
		})
	}
}

func WithMinioAutodetectEndpoint() SubReconcilerOption {
	if _, err := rest.InClusterConfig(); err == nil {
		return WithMinioInClusterEndpoint()
	}
	return WithMinioOutOfClusterEndpoint()
}
