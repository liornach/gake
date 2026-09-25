package logic

import (
	"errors"
	"fmt"
	"gake/utils"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const cProjExpectedCmakeRoot = `cmake_minimum_required(VERSION 3.28)
project(test_project C)

set(CMAKE_C_STANDARD 23)
set(CMAKE_C_STANDARD_REQUIRED True)

enable_testing()

add_subdirectory(./src)
add_subdirectory(./tests)`

func getSourceLines(sources []string) string {
	var lines strings.Builder
	for _, source := range sources {
		lines.WriteString("\n    " + source)
	}

	return lines.String()
}

func getCProjExpectedCmakeSrc(libName string, sources ...string) string {
	return fmt.Sprintf(`cmake_minimum_required(VERSION 3.28)

add_library(%[1]s OBJECT)

target_sources(%[1]s
    PRIVATE%[2]s
)

target_include_directories(%[1]s
    PUBLIC
    ${PROJECT_SOURCE_DIR}/include
)`, libName, getSourceLines(sources))
}

func getCXXProjExpectedCmakeSrc(sources ...string) string {
	return fmt.Sprintf(`cmake_minimum_required(VERSION 3.28)

add_library(Objects OBJECT)

target_sources(Objects
    PUBLIC
    FILE_SET CXX_MODULES FILES%s
)`, getSourceLines(sources))
}

func getCProjExpectedCmakeTests(libName string) string {
	return fmt.Sprintf(`function(AddTest TEST_NAME TEST_SOURCE)
    add_executable(${TEST_NAME}
        ${TEST_SOURCE}
    )

    target_link_libraries(${TEST_NAME}
        PRIVATE %s
    )

    add_test(NAME ${TEST_NAME} COMMAND ${TEST_NAME})
endfunction()
AddTest(Core core.c)
AddTest(Utils utils.c)`, libName)
}

// createStubCopy copies one stub into a fresh per-test temp dir, keeping the stub's
// directory name (the C library is named after it). Go deletes it after the test.
func createStubCopy(t *testing.T, stubName string) string {
	t.Helper()
	testDir, err := thisTestDir()
	if err != nil {
		t.Fatal(err)
	}

	root := filepath.Join(t.TempDir(), stubName)
	stubDir := filepath.Join(testDir, testsAssets, "project-stubs", stubName)
	if err := utils.CopyDir(stubDir, root); err != nil {
		t.Fatal(err)
	}

	return root
}

// createProjInDir builds a project from empty files inside a temp dir named dirName.
func createProjInDir(t *testing.T, dirName string, files ...string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), dirName)
	for _, file := range files {
		path := filepath.Join(root, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}

	return root
}

func readCmake(t *testing.T, dir string) string {
	t.Helper()
	content, err := os.ReadFile(utils.JoinCmakeLists(dir))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

func assertCmakeContent(t *testing.T, dir, expected string) {
	t.Helper()
	got := readCmake(t, dir)
	if got != expected {
		t.Fatalf("%s differs from expected\n--- got ---\n%s\n--- want ---\n%s", utils.JoinCmakeLists(dir), got, expected)
	}
}

func assertDirExists(t *testing.T, dir string) {
	t.Helper()
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected directory %s: %v", dir, err)
	}

	if !info.IsDir() {
		t.Fatalf("%s exists but is not a directory", dir)
	}
}

func assertNoDir(t *testing.T, dir string) {
	t.Helper()
	if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no %s, got err %v", dir, err)
	}
}

// snapshotDir maps every file and directory under root (relative path) to its content.
// Directories are recorded too, so creating an empty directory counts as a change.
func snapshotDir(t *testing.T, root string) map[string]string {
	t.Helper()
	snapshot := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(root, path)
		if d.IsDir() {
			snapshot[rel+"/"] = ""
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		snapshot[rel] = string(content)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return snapshot
}

func assertDirUnchanged(t *testing.T, root string, before map[string]string) {
	t.Helper()
	if after := snapshotDir(t, root); !maps.Equal(before, after) {
		t.Fatalf("%s was modified:\nbefore: %v\nafter:  %v", root, slices.Sorted(maps.Keys(before)), slices.Sorted(maps.Keys(after)))
	}
}

func TestNewCProj(t *testing.T) {
	root := createStubCopy(t, "new-c")

	if err := gake(root, utils.Args{ProjName: testProjName()}); err != nil {
		t.Fatal(err)
	}

	assertCmakeContent(t, root, cProjExpectedCmakeRoot)
	assertCmakeContent(t, utils.JoinPath(root, "src"), getCProjExpectedCmakeSrc("new-c", "core.c", "utils.c"))
	assertCmakeContent(t, utils.JoinPath(root, "tests"), getCProjExpectedCmakeTests("new-c"))
	assertDirExists(t, utils.JoinPath(root, "include"))
}

func TestExistingCProj(t *testing.T) {
	root := createStubCopy(t, "existing-c")
	rootCmakeBefore := readCmake(t, root)

	if err := gake(root, utils.Args{}); err != nil {
		t.Fatal(err)
	}

	assertCmakeContent(t, root, rootCmakeBefore)
	assertCmakeContent(t, utils.JoinPath(root, "src"), getCProjExpectedCmakeSrc("existing-c", "core.c", "utils.c"))
	assertCmakeContent(t, utils.JoinPath(root, "tests"), getCProjExpectedCmakeTests("existing-c"))
	assertDirExists(t, utils.JoinPath(root, "include"))
}

func TestNewMixedProjFails(t *testing.T) {
	root := createStubCopy(t, "new-mixed")
	before := snapshotDir(t, root)

	err := gake(root, utils.Args{ProjName: testProjName()})
	if !errors.Is(err, utils.ErrMixedSources) {
		t.Fatalf("got err %v, want %v", err, utils.ErrMixedSources)
	}

	assertDirUnchanged(t, root, before)
}

func TestNewCProjInvalidLibName(t *testing.T) {
	root := createProjInDir(t, "my lib", "src/a.c", "tests/a.c")
	before := snapshotDir(t, root)

	err := gake(root, utils.Args{ProjName: testProjName()})
	if !errors.Is(err, utils.ErrInvalidLibName) {
		t.Fatalf("got err %v, want %v", err, utils.ErrInvalidLibName)
	}

	assertDirUnchanged(t, root, before)
}

func TestExistingCProjInvalidLibName(t *testing.T) {
	root := createProjInDir(t, "my lib", "src/a.c", "tests/a.c")
	if err := os.WriteFile(utils.JoinCmakeLists(root), []byte("project(x C)"), 0644); err != nil {
		t.Fatal(err)
	}
	before := snapshotDir(t, root)

	err := gake(root, utils.Args{})
	if !errors.Is(err, utils.ErrInvalidLibName) {
		t.Fatalf("got err %v, want %v", err, utils.ErrInvalidLibName)
	}

	assertDirUnchanged(t, root, before)
}

// The library-name rule is C only: C++ keeps "Objects", whatever the directory is called.
func TestNewCXXProjIgnoresLibNameRule(t *testing.T) {
	root := createProjInDir(t, "my lib", "src/a.cppm", "tests/a.cpp")

	if err := gake(root, utils.Args{ProjName: testProjName()}); err != nil {
		t.Fatal(err)
	}

	assertCmakeContent(t, utils.JoinPath(root, "src"), getCXXProjExpectedCmakeSrc("a.cppm"))
	assertNoDir(t, utils.JoinPath(root, "include"))
}

func TestNewProjMissingProjName(t *testing.T) {
	root := createStubCopy(t, "new-c")
	before := snapshotDir(t, root)

	err := gake(root, utils.Args{})
	if !errors.Is(err, utils.ErrMissingProjName) {
		t.Fatalf("got err %v, want %v", err, utils.ErrMissingProjName)
	}

	assertDirUnchanged(t, root, before)
}
