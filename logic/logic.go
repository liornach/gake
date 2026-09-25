package logic

import (
	"gake/utils"
	"os"
)

func RunAtUserRoot() error {
	wd, err := utils.Wd()
	if err != nil {
		return err
	}

	return gake(wd)
}

func gake(root string) error {
	contain, err := utils.ContainCmakeLists(root)
	if err != nil {
		panic(err)
	}

	if !contain {
		projName, err := utils.ProjnameArgv()
		if err != nil {
			return wrapError("failed to get projname from argv", err)
		}

		if err := initNewProj(root, projName); err != nil {
			return wrapError("failed to init new proj", err)
		}
	} else {
		if err := updateExistingProj(root); err != nil {
			return wrapError("failed to update existing project", err)
		}
	}

	return nil
}

func initNewProj(root, projname string) error {
	if err := initNewRootCmake(root, projname); err != nil {
		panic(err)
	}

	if err := initSrcCmake(root); err != nil {
		panic(err)
	}

	if err := initTestCmake(root); err != nil {
		panic(err)
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
	if err := utils.Mkdir(srcDir); err != nil {
		return err
	}

	cppmFiles, err := utils.CollectAllFilesWithExt(srcDir, "cppm")
	if err != nil {
		return err
	}

	content, err := utils.InsertDataToSrcTemplate(cppmFiles)
	if err != nil {
		return err
	}

	return createCmakeAtDir(srcDir, content)
}

func initTestCmake(root string) error {
	tstDir := root + "/tests"
	if err := utils.Mkdir(tstDir); err != nil {
		return err
	}

	tests, err := utils.CollectTests(tstDir)
	if err != nil {
		panic(err)
	}

	content, err := utils.InsertDataToTestsTemplate(tests)
	if err != nil {
		panic(err)
	}

	return createCmakeAtDir(tstDir, content)
}

func initNewRootCmake(root, projname string) error {
	content, err := utils.InsertProjnameToTemplate(projname)
	if err != nil {
		panic(err)
	}

	if err := createCmakeAtDir(root, []byte(content)); err != nil {
		panic(err)
	}

	return nil
}

func createCmakeAtDir(dir string, content []byte) error {
	file := utils.JoinCmakeLists(dir)
	return os.WriteFile(file, content, 0644)
}
