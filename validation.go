package main

import "bytes"

func validateExistingRootCmake(userRoot string) error {
	projname, err := getProjNameFromCmake(userRoot)
	if err != nil {
		return err
	}

	rendered, err := insertProjnameToTemplate(projname)
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

func validateRootDirectoryStructure(root string) error {
	for _, name := range []string{"src", "tests"} {
		exist, err := dirExists(joinPath(root, name))
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
