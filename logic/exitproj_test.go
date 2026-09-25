package logic

import (
	"gake/utils"
	"testing"
)

const cxxProjExpectedCmakeTests = `function(AddTest TEST_NAME TEST_SOURCE)
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

// The existing stub already holds empty src/ and tests/ CMakeLists, so the content is
// what proves they were regenerated.
func TestExistingProk(t *testing.T) {
	root := createStubCopy(t, "existing")
	rootCmakeBefore := readCmake(t, root)

	if err := gake(root, utils.Args{}); err != nil {
		t.Fatal(err)
	}

	assertCmakeContent(t, root, rootCmakeBefore)
	assertCmakeContent(t, utils.JoinPath(root, "src"), newProjExpectedCmakeSrc)
	assertCmakeContent(t, utils.JoinPath(root, "tests"), cxxProjExpectedCmakeTests)
	assertNoDir(t, utils.JoinPath(root, "include"))
}
