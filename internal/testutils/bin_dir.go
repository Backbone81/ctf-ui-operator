package testutils

import (
	"fmt"
	"os"
	"path"
)

// PrependToPath prepends the given directory to the PATH environment variable.
func PrependToPath(directory string) error {
	oldPath := os.Getenv("PATH")
	newPath := fmt.Sprintf("%s:%s", directory, oldPath)
	return os.Setenv("PATH", newPath)
}

// MakeBinDirAvailable adds the bin subdirectory of the current working directory to the PATH environment variable.
// This is helpful when debugging tests, because some IDEs make it hard to extend the PATH environment variable through
// their project configuration.
func MakeBinDirAvailable() error {
	currDir, err := os.Getwd()
	if err != nil {
		return err
	}
	return PrependToPath(path.Join(currDir, "bin"))
}
