package logic

import (
	"errors"
	"fmt"
	"gake/utils"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const testsAssets = "tests-assets"

func testSandboxPath(testDir string) string {
	return utils.JoinPath(testDir, ".tmp")
}

func copyAssetsToTestSandbox(testAssetsDir, testDir string) (string, error) {
	sandboxDir := testSandboxPath(testDir)
	if err := utils.CopyDir(testAssetsDir, sandboxDir); err != nil {
		return "", err
	}

	return sandboxDir, nil
}

func thisTestDir() (string, error) {
	_, thisFilename, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("failed to call runtime.Caler")
	}

	testsDir := filepath.Dir(thisFilename)
	return testsDir, nil
}

func initTestSandbox() (string, error) {
	testsDir, err := thisTestDir()
	if err != nil {
		panic(err)
	}

	assetsDir := utils.JoinPath(testsDir, testsAssets)
	sandboxDir, err := copyAssetsToTestSandbox(assetsDir, testsDir)
	if err != nil {
		panic(err)
	}

	return sandboxDir, nil
}

func testProjName() string {
	return "test_project"
}

func newProjSandboxRootPath(sandboxDir string) string {
	return utils.JoinPath(sandboxDir, "project-stubs/new")
}

const newProjTestName = "test_project"

var newProjExpectedCmakeRoot = fmt.Sprintf(`cmake_minimum_required(VERSION 3.28)
project(%s CXX)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED True)

enable_testing()

add_subdirectory(./src)
add_subdirectory(./tests)`, newProjTestName)

const newProjExpectedCmakeSrc = `cmake_minimum_required(VERSION 3.28)

add_library(Objects OBJECT)

target_sources(Objects
    PUBLIC
    FILE_SET CXX_MODULES FILES
    core.cppm
    fuck_you.cppm
    helper.cppm
    utils.cppm
)`

func TestNewProj(t *testing.T) {
	sandboxDir, err := initTestSandbox()
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(sandboxDir)

	sandboxRoot := newProjSandboxRootPath(sandboxDir)
	if err := initNewProj(sandboxRoot, "testproj"); err != nil {
		t.Fatal(err)
	}

	if err = initNewProj(sandboxRoot, testProjName()); err != nil {
		t.Fatal(err)
	}

	res, err := os.ReadFile(utils.JoinCmakeLists(sandboxRoot))
	if err != nil {
		t.Fatal(err)
	}

	if newProjExpectedCmakeRoot != string(res) {
		t.Fatal("CmakeLists.txt at root is different than expected")
	}

	res, err = os.ReadFile(utils.JoinCmakeLists(utils.JoinPath(sandboxRoot, "src")))
	if err != nil {
		t.Fatal(err)
	}

	if newProjExpectedCmakeSrc != string(res) {
		t.Fatal("CmakeLists.txt at src dir is different than expected")
	}

	res, err = os.ReadFile(utils.JoinCmakeLists(utils.JoinPath(sandboxRoot, "tests")))
	if err != nil {
		t.Fatal(err)
	}

	const expected = `function(AddTest TEST_NAME TEST_SOURCE)
    add_executable(${TEST_NAME}
        ${TEST_SOURCE}
    )

    target_link_libraries(${TEST_NAME}
        PRIVATE Objects
    )

    add_test(NAME ${TEST_NAME} COMMAND ${TEST_NAME})
endfunction()
AddTest(Core core.cpp)
AddTest(FuckYou fuck_you.cpp)
AddTest(Helper helper.cpp)
AddTest(Utils utils.cpp)`

	if expected != string(res) {
		t.Fatal("CMakeLists.txt at tests dir is different than expected")
	}
}
