package ctfd

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// CTFdEndpointStrategy describes the way to get the endpoint of CTFd. As we need to differentiate between running
// in-cluster and running out-of-cluster, both strategies need to implement this interface.
//
//nolint:iface // CTFdEndpoint strategy and MinioEndpointStrategy have identical methods by design.
type CTFdEndpointStrategy interface {
	GetEndpoint(ctx context.Context, ctfd *v1alpha1.CTFd) (string, error)
}

// InClusterCTFdEndpointStrategy returns an endpoint which is the service name and the port for in-cluster usage.
type InClusterCTFdEndpointStrategy struct{}

var _ CTFdEndpointStrategy = (*InClusterCTFdEndpointStrategy)(nil)

func (s *InClusterCTFdEndpointStrategy) GetEndpoint(ctx context.Context, ctfd *v1alpha1.CTFd) (string, error) {
	return fmt.Sprintf("http://%s.%s:80", ctfd.Name, ctfd.Namespace), nil
}

// OutOfClusterCTFdEndpointStrategy port forwards the CTFd service to the local host and returns an endpoint with
// that forwarded port. The local port is a random free port.
type OutOfClusterCTFdEndpointStrategy struct {
	servicePortForwarder *utils.ServicePortForwarder
}

var _ CTFdEndpointStrategy = (*OutOfClusterCTFdEndpointStrategy)(nil)

func (s *OutOfClusterCTFdEndpointStrategy) GetEndpoint(ctx context.Context, ctfd *v1alpha1.CTFd) (string, error) {
	localPort, err := s.servicePortForwarder.PortForward(
		ctx,
		types.NamespacedName{
			Namespace: ctfd.Namespace,
			Name:      ctfd.Name,
		},
		intstr.FromString("http"),
	)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://127.0.0.1:%d", localPort), nil
}

type CTFdEndpointSetter interface {
	SetCTFdEndpoint(ctfdEndpointStrategy CTFdEndpointStrategy)
}

type SubReconcilerOption func(subReconciler any)

func WithCTFdInClusterEndpoint() SubReconcilerOption {
	return func(subReconciler any) {
		endpointSetter, ok := subReconciler.(CTFdEndpointSetter)
		if !ok {
			panic("this option requires the sub reconciler to implement the CTFdEndpointSetter interface")
		}
		endpointSetter.SetCTFdEndpoint(&InClusterCTFdEndpointStrategy{})
	}
}

type ClientGetter interface {
	GetClient() client.Client
}

func WithCTFdOutOfClusterEndpoint() SubReconcilerOption {
	return func(subReconciler any) {
		clientGetter, ok := subReconciler.(ClientGetter)
		if !ok {
			panic("this option requires the sub reconciler to implement the ClientGetter interface")
		}
		endpointSetter, ok := subReconciler.(CTFdEndpointSetter)
		if !ok {
			panic("this option requires the sub reconciler to implement the CTFdEndpointSetter interface")
		}
		endpointSetter.SetCTFdEndpoint(&OutOfClusterCTFdEndpointStrategy{
			servicePortForwarder: utils.NewServicePortForwarder(clientGetter.GetClient()),
		})
	}
}

func WithCTFdAutodetectEndpoint() SubReconcilerOption {
	if _, err := rest.InClusterConfig(); err == nil {
		return WithCTFdInClusterEndpoint()
	}
	return WithCTFdOutOfClusterEndpoint()
}
