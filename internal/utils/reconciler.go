package utils

import (
	"context"
	"reflect"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler is a generalization of a top level reconciler. The type parameter should be a pointer to the kubernetes
// data type this reconciler will reconcile. Every reconcile event is then forwarded to all sub reconcilers.
type Reconciler[T client.Object] struct {
	client         client.Client
	subReconcilers []SubReconciler[T]
	newObj         func() T
}

// NewReconciler creates a new reconciler instance. The reconciler is initialized with the given client and applies
// the provided options to the reconciler.
func NewReconciler[T client.Object](client client.Client, newObj func() T, options ...ReconcilerOption[T]) *Reconciler[T] {
	result := &Reconciler[T]{
		client: client,
		newObj: newObj,
	}
	for _, option := range options {
		option(result)
	}
	return result
}

// SetupWithManager registers all enabled sub-reconcilers with the given manager.
func (r *Reconciler[T]) SetupWithManager(mgr ctrl.Manager) error {
	ctrlBuilder := ctrl.NewControllerManagedBy(mgr).
		For(r.newObj())
	for _, subReconciler := range r.subReconcilers {
		ctrlBuilder = subReconciler.SetupWithManager(ctrlBuilder)
	}
	return ctrlBuilder.Complete(r)
}

func (r *Reconciler[T]) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	obj, err := r.getObject(ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}
	if reflect.ValueOf(obj).IsZero() {
		// The resource was deleted.
		return ctrl.Result{}, nil
	}

	for _, subReconciler := range r.subReconcilers {
		result, err := subReconciler.Reconcile(ctx, obj)
		if err != nil || !result.IsZero() {
			return result, IgnoreConflict(err)
		}
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler[T]) getObject(ctx context.Context, req ctrl.Request) (T, error) {
	result := r.newObj()
	if err := r.client.Get(ctx, req.NamespacedName, result); err != nil {
		var zero T
		return zero, client.IgnoreNotFound(err)
	}
	return result, nil
}

func (r *Reconciler[T]) AppendSubReconciler(subReconciler SubReconciler[T]) {
	r.subReconcilers = append(r.subReconcilers, subReconciler)
}

func (r *Reconciler[T]) GetClient() client.Client {
	return r.client
}

// SubReconciler is the interface all sub-reconcilers need to implement.
type SubReconciler[T client.Object] interface {
	Reconcile(ctx context.Context, obj T) (ctrl.Result, error)
	SetupWithManager(builder *builder.Builder) *builder.Builder
}

// ReconcilerOption is an option which can be applied to the reconciler.
type ReconcilerOption[T client.Object] func(reconciler *Reconciler[T])
