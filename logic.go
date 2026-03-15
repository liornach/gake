package main

import (
	"os"
)

func runAtUserRoot() error {
	wd, err := wd()
	if err != nil {
		return err
	}

	return gake(wd)
}

func gake(root string) error {
	contain, err := containCmakeLists(root)
	if err != nil {
		return err
	}

	if !contain {
		return initNewProj(root)

	}

	return updateExistingProj(root)
}

func initNewProj(root string) error {
	projName, err := projnameArgv()
	if err != nil {
		return err
	}

	if err := initNewRootCmake(root, projName); err != nil {
		return err
	}

	if err := initSrcCmake(root); err != nil {
		return err
	}

	if err := initTestCmake(root); err != nil {
		return err
	}

	return nil
}

func updateExistingProj(root string) error {
	if err := initSrcCmake(root); err != nil {
		return err
	}

	if err := initTestCmake(root); err != nil {
		return err
	}

	return nil
}

func initSrcCmake(root string) error {
	srcDir := root + "/src"
	if err := mkdir(srcDir); err != nil {
		return err
	}

	cppmFiles, err := collectAllFilesWithExt(srcDir, "cppm")
	if err != nil {
		return err
	}

	content, err := insertDataToSrcTemplate(cppmFiles)
	if err != nil {
		return err
	}

	return createCmakeAtDir(srcDir, content)
}

func initTestCmake(root string) error {
	tstDir := root + "/tests"
	if err := mkdir(tstDir); err != nil {
		return err
	}

	tests, err := collectTests(tstDir)
	if err != nil {
		return err
	}

	content, err := insertDataToTestsTemplate(tests)
	if err != nil {
		return err
	}

	return createCmakeAtDir(tstDir, content)
}

func initNewRootCmake(root, projname string) error {
	content, err := insertProjnameToTemplate(projname)
	if err != nil {
		return err
	}

	return createCmakeAtDir(root, []byte(content))
}

func createCmakeAtDir(dir string, content []byte) error {
	file := joinCmakeLists(dir)
	return os.WriteFile(file, content, 0644)
}
