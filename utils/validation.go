package utils

import (
	"bytes"
	"fmt"
)

type notMadeByUsError struct {
}

func (notMadeByUsError) Error() string {
	return "Looks like the existing CMakeLists.txt file was not made by us, since it is not matching our expected template"
}

func validateExistingRootCmake(userRoot string) error {
	projname, err := readProjNameFromCmake(userRoot)
	if err != nil {
		return err
	}

	rendered, err := InsertProjnameToTemplate(projname)
	if err != nil {
		return err
	}

	existing, err := readExistingCmake(userRoot)
	if err != nil {
		return err
	}

	if !bytes.Equal([]byte(rendered), existing) {
		return notMadeByUsError{}
	}

	return nil
}

type missingDirError struct {
	dir string
}

func (e missingDirError) Error() string {
	return fmt.Sprintf("expected directory structure missing: %s", e.dir)
}

func validateRootDirectoryStructure(root string) error {
	for _, name := range []string{"src", "tests"} {
		exist, err := dirExists(JoinPath(root, name))
		if err != nil {
			return err
		}

		if !exist {
			return missingDirError{
				dir: name,
			}
		}
	}
	return nil
}
