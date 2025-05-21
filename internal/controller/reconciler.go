package controller

import (
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/backbone81/ctf-ui-operator/internal/controller/ctfd"
	"github.com/backbone81/ctf-ui-operator/internal/controller/mariadb"
	"github.com/backbone81/ctf-ui-operator/internal/controller/redis"
)

// Reconciler is the main reconciler of this operator. It is responsible for registering and running all
// top level reconcilers.
type Reconciler struct {
	client         client.Client
	subReconcilers []SubReconciler
}

// NewReconciler creates a new reconciler instance. The reconciler is initialized with the given client and applies
// the provided options to the reconciler.
func NewReconciler(client client.Client, options ...ReconcilerOption) *Reconciler {
	result := &Reconciler{
		client: client,
	}
	for _, option := range options {
		option(result)
	}
	return result
}

// SetupWithManager registers all enabled sub-reconcilers with the given manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	for _, subReconciler := range r.subReconcilers {
		if err := subReconciler.SetupWithManager(mgr); err != nil {
			return err
		}
	}
	return nil
}

// SubReconciler is the interface all sub-reconcilers need to implement.
type SubReconciler interface {
	reconcile.Reconciler
	SetupWithManager(mgr ctrl.Manager) error
}

// ReconcilerOption is an option which can be applied to the reconciler.
type ReconcilerOption func(reconciler *Reconciler)

// WithDefaultReconcilers returns a reconciler option which enables the default sub-reconcilers.
func WithDefaultReconcilers(recorder record.EventRecorder) ReconcilerOption {
	return func(reconciler *Reconciler) {
		WithRedisReconciler()(reconciler)
		WithMariaDBReconciler()(reconciler)
	}
}

// WithCTFdReconciler returns a reconciler option which enables the CTFd sub-reconciler.
func WithCTFdReconciler(recorder record.EventRecorder) ReconcilerOption {
	return func(reconciler *Reconciler) {
		reconciler.subReconcilers = append(
			reconciler.subReconcilers,
			ctfd.NewReconciler(reconciler.client, ctfd.WithDefaultReconcilers(recorder)),
		)
	}
}

// WithMariaDBReconciler returns a reconciler option which enables the MariaDB sub-reconciler.
func WithMariaDBReconciler() ReconcilerOption {
	return func(reconciler *Reconciler) {
		reconciler.subReconcilers = append(
			reconciler.subReconcilers,
			mariadb.NewReconciler(reconciler.client, mariadb.WithDefaultReconcilers()),
		)
	}
}

// WithRedisReconciler returns a reconciler option which enables the Redis sub-reconciler.
func WithRedisReconciler() ReconcilerOption {
	return func(reconciler *Reconciler) {
		reconciler.subReconcilers = append(
			reconciler.subReconcilers,
			redis.NewReconciler(reconciler.client, redis.WithDefaultReconcilers()),
		)
	}
}
