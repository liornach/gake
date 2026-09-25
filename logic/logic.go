package logic

import (
	"errors"
	"flag"
	"gake/utils"
	"io"
	"os"
)

func RunAtUserRoot() error {
	wd, err := utils.Wd()
	if err != nil {
		return err
	}

	return runGake(wd, os.Args[1:], os.Stdout)
}

// runGake is the command line: parse argv, print usage on -h, otherwise run gake in root.
func runGake(root string, argv []string, out io.Writer) error {
	args, err := utils.ParseArgs(argv)
	if errors.Is(err, flag.ErrHelp) {
		utils.WriteUsage(out)
		return nil
	}
	if err != nil {
		return wrapError("failed to parse arguments", err)
	}

	return gake(root, args)
}

func gake(root string, args utils.Args) error {
	isExistingProj, err := utils.ContainCmakeLists(root)
	if err != nil {
		return err
	}

	if isExistingProj {
		return wrapError("failed to update existing project", updateExistingProj(root, args))
	}

	return wrapError("failed to init new proj", createNewProj(root, args))
}

// initNewProj creates a new project without a --lang flag. Kept for TestNewProj.
func initNewProj(root, projname string) error {
	return createNewProj(root, utils.Args{ProjName: projname})
}

// createNewProj validates everything first, then writes the root, src and tests CMakeLists.
func createNewProj(root string, args utils.Args) error {
	if args.ProjName == "" {
		return utils.ErrMissingProjName
	}

	lang, err := resolveNewProjLanguage(root, args)
	if err != nil {
		return err
	}

	profile, err := utils.GetProfile(lang, root)
	if err != nil {
		return err
	}

	rootCmake, err := renderRootCmake(root, profile, args.ProjName)
	if err != nil {
		return err
	}

	dirCmakes, err := renderProjDirCmakes(root, profile)
	if err != nil {
		return err
	}

	return writeProj(root, profile, append(dirCmakes, rootCmake))
}

// updateExistingProj validates everything first, then regenerates the src and tests CMakeLists.
func updateExistingProj(root string, args utils.Args) error {
	lang, err := resolveExistingProjLanguage(root, args)
	if err != nil {
		return err
	}

	profile, err := utils.GetProfile(lang, root)
	if err != nil {
		return err
	}

	dirCmakes, err := renderProjDirCmakes(root, profile)
	if err != nil {
		return err
	}

	return writeProj(root, profile, dirCmakes)
}

// cmakeFile is a rendered CMakeLists.txt waiting to be written into dir.
type cmakeFile struct {
	dir     string
	content []byte
}

func renderRootCmake(root string, profile utils.Profile, projName string) (cmakeFile, error) {
	content, err := utils.RenderRootCmake(profile, projName)
	return cmakeFile{dir: root, content: content}, err
}

// renderProjDirCmakes renders the src and tests CMakeLists without touching the disk.
func renderProjDirCmakes(root string, profile utils.Profile) ([]cmakeFile, error) {
	srcCmake, err := renderSrcCmake(root, profile)
	if err != nil {
		return nil, err
	}

	testsCmake, err := renderTestsCmake(root, profile)
	if err != nil {
		return nil, err
	}

	return []cmakeFile{srcCmake, testsCmake}, nil
}

func renderSrcCmake(root string, profile utils.Profile) (cmakeFile, error) {
	srcDir := utils.GetSrcDir(root)
	sources, err := utils.CollectAllFilesWithExt(srcDir, profile.SrcExt)
	if err != nil {
		return cmakeFile{}, err
	}

	content, err := utils.RenderSrcCmake(profile, sources)
	return cmakeFile{dir: srcDir, content: content}, err
}

func renderTestsCmake(root string, profile utils.Profile) (cmakeFile, error) {
	testsDir := utils.GetTestsDir(root)
	tests, err := utils.CollectTests(testsDir, profile.TestExt)
	if err != nil {
		return cmakeFile{}, err
	}

	content, err := utils.RenderTestsCmake(profile, tests)
	return cmakeFile{dir: testsDir, content: content}, err
}

// writeProj is the only step that writes: the profile's extra dirs and every rendered CMakeLists.
func writeProj(root string, profile utils.Profile, cmakes []cmakeFile) error {
	for _, dir := range profile.ExtraDirs {
		if err := utils.Mkdir(utils.JoinPath(root, dir)); err != nil {
			return err
		}
	}

	for _, cmake := range cmakes {
		if err := createCmakeAtDir(cmake.dir, cmake.content); err != nil {
			return err
		}
	}

	return nil
}

func createCmakeAtDir(dir string, content []byte) error {
	if err := utils.Mkdir(dir); err != nil {
		return err
	}

	file := utils.JoinCmakeLists(dir)
	return os.WriteFile(file, content, 0644)
}
