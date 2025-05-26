package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServicePortForwarder creates port forwards to given services and allocates a random local port for that port forward.
// It keeps a list of all active port forwards. It allows for on-the-fly port-forwards when needing access to services
// inside the kubernetes cluster, when running the operator locally.
type ServicePortForwarder struct {
	client             client.Client
	mutex              sync.Mutex
	activePortForwards map[string]*portforward.PortForwarder
}

// NewServicePortForwarder creates a new instance of ServicePortForwarder.
func NewServicePortForwarder(client client.Client) *ServicePortForwarder {
	return &ServicePortForwarder{
		client:             client,
		activePortForwards: make(map[string]*portforward.PortForwarder),
	}
}

// PortForward creates a port forwarding for the given service and port. It returns the local port the port forward
// is listening for connections. If there is already a port forwarding active for the given service, the local port
// of the existing port forwarding is returned.
//
//nolint:funlen
func (f *ServicePortForwarder) PortForward(ctx context.Context, serviceName types.NamespacedName, servicePort intstr.IntOrString) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	key := fmt.Sprintf("%s:%s", serviceName.String(), servicePort.String())
	if activePortForward, exists := f.activePortForwards[key]; exists {
		// We already have a port forwarding active for the given object.
		return f.getLocalPort(activePortForward)
	}

	// We need to fetch the service and retrieve the target port the service is forwarding to.
	service, err := f.getService(ctx, serviceName)
	if err != nil {
		return 0, err
	}
	targetPort, err := f.getServiceTargetPort(service, servicePort)
	if err != nil {
		return 0, err
	}

	// We fetch one pod which is targeted by the service and retrieve the port on the pod.
	pod, err := f.getServicePod(ctx, service)
	if err != nil {
		return 0, err
	}
	podPort, err := f.getPodPort(pod, targetPort)
	if err != nil {
		return 0, err
	}

	dialer, err := f.getPodDialer(pod)
	if err != nil {
		return 0, fmt.Errorf("getting dialer for port forward: %w", err)
	}

	// Create port forwarder
	stopChan := make(chan struct{})
	readyChan := make(chan struct{})
	newPortForward, err := portforward.New(
		dialer,
		[]string{fmt.Sprintf(":%d", podPort)},
		stopChan,
		readyChan,
		io.Discard, // we are not interested in the log output of the port forward
		io.Discard, // we are not interested in the log output of the port forward
	)
	if err != nil {
		return 0, fmt.Errorf("creating port forwarder: %w", err)
	}

	// Start port forwarder
	f.activePortForwards[key] = newPortForward
	go f.runPortForwarder(ctx, key, newPortForward)

	// Wait for the port forward to be ready or the context to be done.
	select {
	case <-ctx.Done():
		newPortForward.Close()
		return 0, errors.New("the context was canceled while waiting for the port forward to become ready")
	case <-readyChan:
	}
	localPort, err := f.getLocalPort(newPortForward)
	ctrl.LoggerFrom(ctx).Info(
		"Port forwarding Service",
		"service-namespace", serviceName.Namespace,
		"service-name", serviceName.Name,
		"service-port", servicePort.String(),
		"local-port", localPort,
	)
	return localPort, err
}

func (f *ServicePortForwarder) getServiceTargetPort(service *corev1.Service, port intstr.IntOrString) (intstr.IntOrString, error) {
	for _, currPort := range service.Spec.Ports {
		switch port.Type {
		case intstr.Int:
			if currPort.Port == port.IntVal {
				return currPort.TargetPort, nil
			}
		case intstr.String:
			if currPort.Name == port.StrVal {
				return currPort.TargetPort, nil
			}
		default:
			return intstr.IntOrString{}, errors.New("unknown service port type")
		}
	}
	return intstr.IntOrString{}, errors.New("port not found on service")
}

func (f *ServicePortForwarder) getPodPort(pod *corev1.Pod, port intstr.IntOrString) (int32, error) {
	for _, container := range pod.Spec.Containers {
		for _, currPort := range container.Ports {
			switch port.Type {
			case intstr.Int:
				if currPort.ContainerPort == port.IntVal {
					return currPort.ContainerPort, nil
				}
			case intstr.String:
				if currPort.Name == port.StrVal {
					return currPort.ContainerPort, nil
				}
			default:
				return 0, errors.New("unknown pod port type")
			}
		}
	}
	return 0, errors.New("port not found on pod")
}

func (f *ServicePortForwarder) getPodDialer(pod *corev1.Pod) (httpstream.Dialer, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("getting kubernetes config: %w", err)
	}

	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, fmt.Errorf("creating round tripper for spdy: %w", err)
	}

	target, err := url.JoinPath(config.Host, config.APIPath, "/api/v1/namespaces", pod.Namespace, "/pods", pod.Name, "/portforward")
	if err != nil {
		return nil, fmt.Errorf("constructing target URL for dialer: %w", err)
	}
	targetUrl, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parsing target URL for dialer: %w", err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, targetUrl)
	return dialer, nil
}

func (f *ServicePortForwarder) getService(ctx context.Context, serviceName types.NamespacedName) (*corev1.Service, error) {
	var service corev1.Service
	if err := f.client.Get(ctx, serviceName, &service); err != nil {
		return nil, err
	}
	return &service, nil
}

func (f *ServicePortForwarder) getServicePod(ctx context.Context, service *corev1.Service) (*corev1.Pod, error) {
	var pods corev1.PodList
	if err := f.client.List(
		ctx,
		&pods,
		client.InNamespace(service.Namespace),
		client.MatchingLabels(service.Spec.Selector),
	); err != nil {
		return nil, err
	}

	// We are looking for a pod which is running and ready.
	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				return &pod, nil
			}
		}
	}
	return nil, errors.New("no ready pods found for service")
}

func (f *ServicePortForwarder) getLocalPort(portForward *portforward.PortForwarder) (int, error) {
	ports, err := portForward.GetPorts()
	if err != nil {
		return 0, fmt.Errorf("getting ports of port forward: %w", err)
	}
	return int(ports[0].Local), nil
}

func (f *ServicePortForwarder) runPortForwarder(ctx context.Context, key string, fw *portforward.PortForwarder) {
	if err := fw.ForwardPorts(); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed to execute port forward")
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	delete(f.activePortForwards, key)
}
