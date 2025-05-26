package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"sync"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	ctrl "sigs.k8s.io/controller-runtime"
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
func (f *ServicePortForwarder) PortForward(ctx context.Context, serviceName types.NamespacedName, servicePort int) (string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if activePortForward, exists := f.activePortForwards[serviceName.String()]; exists {
		// We already have a port forwarding active for the given object.
		return f.getLocalPort(activePortForward)
	}

	dialer, err := f.getDialer(ctx, serviceName)
	if err != nil {
		return "", fmt.Errorf("getting dialer for port forward: %w", err)
	}

	// Create port forwarder
	stopChan := make(chan struct{})
	readyChan := make(chan struct{})
	newPortForward, err := portforward.New(
		dialer,
		[]string{fmt.Sprintf(":%d", servicePort)},
		stopChan,
		readyChan,
		io.Discard,
		io.Discard,
	)
	if err != nil {
		return "", fmt.Errorf("creating port forwarder: %w", err)
	}

	// Start port forwarder
	f.activePortForwards[serviceName.String()] = newPortForward
	go f.runPortForwarder(ctx, serviceName, newPortForward)

	// Wait for the port forward to be ready or the context to be done.
	select {
	case <-ctx.Done():
		newPortForward.Close()
		return "", errors.New("the context was canceled while waiting for the port forward to become ready")
	case <-stopChan:
	case <-readyChan:
	}
	return f.getLocalPort(newPortForward)
}

func (f *ServicePortForwarder) getDialer(ctx context.Context, serviceName types.NamespacedName) (httpstream.Dialer, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("getting kubernetes config: %w", err)
	}

	pod, err := f.getServicePod(ctx, serviceName)
	if err != nil {
		return nil, err
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

func (f *ServicePortForwarder) getServicePod(ctx context.Context, serviceName types.NamespacedName) (*corev1.Pod, error) {
	var service corev1.Service
	if err := f.client.Get(ctx, serviceName, &service); err != nil {
		return nil, err
	}

	var pods corev1.PodList
	if err := f.client.List(ctx, &pods, client.InNamespace(service.Namespace), client.MatchingLabels(service.Spec.Selector)); err != nil {
		return nil, err
	}

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
	return nil, fmt.Errorf("no ready pods found for service")
}

func (f *ServicePortForwarder) getLocalPort(portForward *portforward.PortForwarder) (string, error) {
	ports, err := portForward.GetPorts()
	if err != nil {
		return "", fmt.Errorf("getting ports of port forward: %w", err)
	}
	return strconv.Itoa(int(ports[0].Local)), nil
}

func (f *ServicePortForwarder) runPortForwarder(ctx context.Context, serviceName types.NamespacedName, fw *portforward.PortForwarder) {
	if err := fw.ForwardPorts(); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed to execute port forward")
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	delete(f.activePortForwards, serviceName.String())
}
