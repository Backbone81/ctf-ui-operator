package utils

import "k8s.io/apimachinery/pkg/api/errors"

// IgnoreConflict hides a conflict error be returning nil. This helps in cleaning up the operator log from errors
// which cannot be acted upon and are expected to occur.
func IgnoreConflict(err error) error {
	if errors.IsConflict(err) {
		return nil
	}
	return err
}
