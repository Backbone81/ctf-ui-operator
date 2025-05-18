package testutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// MoveToProjectRoot changes the current working directory to the root of the project. This is helpful for running
// tests, which are executed with the working directory set to the directory the test resides in. This makes it
// difficult to reference files from other directories. MoveToProjectRoot moves up in the directory tree until it
// finds the go.mod.
func MoveToProjectRoot() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			if err := os.Chdir(currentDir); err != nil {
				return fmt.Errorf("changing current directory: %w", err)
			}
			return nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return errors.New("go.mod not found in any parent directory")
		}
		currentDir = parentDir
	}
}
