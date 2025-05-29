package utils

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewLoggingClient(client client.Client) *LoggingClient {
	return &LoggingClient{
		client: client,
	}
}

// LoggingClient is a Kubernetes client which is creating log entries for every modifying action.
type LoggingClient struct {
	client client.Client
}

// LoggingClient implements client.Client.
var _ client.Client = (*LoggingClient)(nil)

func (l *LoggingClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return l.client.Get(ctx, key, obj, opts...)
}

func (l *LoggingClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return l.client.List(ctx, list, opts...)
}

func (l *LoggingClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	l.logAction(ctx, "Creating", obj)
	return l.client.Create(ctx, obj, opts...)
}

func (l *LoggingClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	l.logAction(ctx, "Deleting", obj)
	return l.client.Delete(ctx, obj, opts...)
}

func (l *LoggingClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	l.logAction(ctx, "Updating", obj)
	return l.client.Update(ctx, obj, opts...)
}

func (l *LoggingClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	l.logAction(ctx, "Patching", obj)
	return l.client.Patch(ctx, obj, patch, opts...)
}

func (l *LoggingClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	l.logAction(ctx, "Deleting all of", obj)
	return l.client.DeleteAllOf(ctx, obj, opts...)
}

func (l *LoggingClient) logAction(ctx context.Context, action string, obj client.Object) {
	controllerruntime.LoggerFrom(ctx).Info(fmt.Sprintf(
		"%s %s",
		action,
		getKind(l.client.Scheme(), obj),
	),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)
}

func (l *LoggingClient) Status() client.SubResourceWriter {
	return &LoggingSubResourceWriter{
		client:      l.client.Status(),
		scheme:      l.client.Scheme(),
		subresource: "status",
	}
}

func (l *LoggingClient) SubResource(subResource string) client.SubResourceClient {
	subResourceClient := l.client.SubResource(subResource)
	return &LoggingSubResourceClient{
		LoggingSubResourceWriter: LoggingSubResourceWriter{
			client:      subResourceClient,
			scheme:      l.client.Scheme(),
			subresource: subResource,
		},
		client: subResourceClient,
	}
}

func (l *LoggingClient) Scheme() *runtime.Scheme {
	return l.client.Scheme()
}

func (l *LoggingClient) RESTMapper() meta.RESTMapper {
	return l.client.RESTMapper()
}

func (l *LoggingClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return l.client.GroupVersionKindFor(obj)
}

func (l *LoggingClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return l.client.IsObjectNamespaced(obj)
}

type LoggingSubResourceWriter struct {
	client      client.SubResourceWriter
	scheme      *runtime.Scheme
	subresource string
}

// LoggingSubResourceWriter implements client.SubResourceWriter.
var _ client.SubResourceWriter = (*LoggingSubResourceWriter)(nil)

func (l *LoggingSubResourceWriter) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	l.logAction(ctx, "Creating", obj)
	return l.client.Create(ctx, obj, subResource, opts...)
}

func (l *LoggingSubResourceWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	l.logAction(ctx, "Updating", obj)
	return l.client.Update(ctx, obj, opts...)
}

func (l *LoggingSubResourceWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	l.logAction(ctx, "Patching", obj)
	return l.client.Patch(ctx, obj, patch, opts...)
}

func (l *LoggingSubResourceWriter) logAction(ctx context.Context, action string, obj client.Object) {
	controllerruntime.LoggerFrom(ctx).Info(fmt.Sprintf(
		"%s %s of %s",
		action,
		l.subresource,
		getKind(l.scheme, obj),
	),
		"name", obj.GetName(),
		"namespace", obj.GetNamespace(),
	)
}

type LoggingSubResourceClient struct {
	LoggingSubResourceWriter
	client client.SubResourceClient
}

// LoggingSubResourceClient implements client.SubResourceClient.
var _ client.SubResourceClient = (*LoggingSubResourceClient)(nil)

func (l *LoggingSubResourceClient) Get(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceGetOption) error {
	return l.client.Get(ctx, obj, subResource, opts...)
}

func getKind(scheme *runtime.Scheme, obj client.Object) string {
	if obj.GetObjectKind().GroupVersionKind().Kind != "" {
		return obj.GetObjectKind().GroupVersionKind().Kind
	}
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil || len(gvks) < 1 {
		return fmt.Sprintf("%T", obj)
	}
	return gvks[0].Kind
}
