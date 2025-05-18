package testutils

// DoNotDeleteFinalizerName provides a name for a finalizer which is not used by the reconcilers themselves. This finalizer
// is used to prevent a resource from being deleted immediately when you want to test situations where you need to
// inspect the behavior for deletion.
var DoNotDeleteFinalizerName = "ctf.backbone81/do-not-delete"
