package ctfd

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/backbone81/ctf-ui-operator/api/v1alpha1"
	"github.com/backbone81/ctf-ui-operator/internal/utils"
)

// EndpointStrategy describes the way to get the endpoint of some component. As we need to differentiate between running
// in-cluster and running out-of-cluster, both strategies need to implement this interface.
type EndpointStrategy interface {
	GetEndpoint(ctx context.Context, ctfd *v1alpha1.CTFd) (string, error)
}

// InClusterEndpointStrategy returns an endpoint which is the service name and the port for in-cluster usage.
type InClusterEndpointStrategy struct{}

var _ EndpointStrategy = (*InClusterEndpointStrategy)(nil)

func (s *InClusterEndpointStrategy) GetEndpoint(ctx context.Context, ctfd *v1alpha1.CTFd) (string, error) {
	return fmt.Sprintf("http://%s.%s:80", ctfd.Name, ctfd.Namespace), nil
}

// OutOfClusterEndpointStrategy port forwards the CTFd service to the local host and returns an endpoint with
// that forwarded port. The local port is a random free port.
type OutOfClusterEndpointStrategy struct {
	servicePortForwarder *utils.ServicePortForwarder
}

var _ EndpointStrategy = (*OutOfClusterEndpointStrategy)(nil)

func (s *OutOfClusterEndpointStrategy) GetEndpoint(ctx context.Context, ctfd *v1alpha1.CTFd) (string, error) {
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
