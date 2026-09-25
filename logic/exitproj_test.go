package logic

import (
	"fmt"
	"gake/utils"
	"os"
	"path/filepath"
	"testing"
)

func validateContainingCmake(path, errMsg string, t *testing.T) {
	if exist, err := utils.IsFileExist(path); err != nil {
		panic(err)
	} else if !exist {
		t.Fatal(errMsg)
	}
}

func TestExistingProk(t *testing.T) {
	sandboxDir, err := initTestSandbox()
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(sandboxDir)

	err = updateExistingProj(sandboxDir)
	if err != nil {
		panic(err)
	}

	srcCmake := filepath.Join(sandboxDir, "src", "CMakeLists.txt")
	validateContainingCmake(srcCmake, fmt.Sprintf("CMakeLists.txt in src direcotry is not exist (file %s not found)", srcCmake), t)
	testCmake := filepath.Join(sandboxDir, "tests", "CMakeLists.txt")
	validateContainingCmake(testCmake, fmt.Sprintf("CMakeLists.txt in tests direcotry is not exist (file %s not found)", srcCmake), t)
}
