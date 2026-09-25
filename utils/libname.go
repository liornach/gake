package utils

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
)

var ErrInvalidLibName = errors.New("project directory name is not a valid library name (allowed: letters, digits, _ and -)")

var libNameRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// GetLibName returns the C library name for a project: the name of its root directory.
func GetLibName(root string) (string, error) {
	name := filepath.Base(root)
	if !libNameRe.MatchString(name) {
		return "", fmt.Errorf("%w: %q", ErrInvalidLibName, name)
	}

	return name, nil
}
