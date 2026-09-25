package logic

import (
	"errors"
	"gake/utils"
	"testing"
)

// getArgsWithLang builds gake's args; an empty langFlag means --lang was not passed.
func getArgsWithLang(langFlag string) utils.Args {
	return utils.Args{ProjName: testProjName(), Lang: langFlag, IsLangSet: langFlag != ""}
}

func assertIncludeDirMatchesLanguage(t *testing.T, root string, lang utils.Language) {
	t.Helper()
	includeDir := utils.JoinPath(root, "include")
	if lang == utils.LangC {
		assertDirExists(t, includeDir)
		return
	}

	assertNoDir(t, includeDir)
}

// How --lang combines with the project, for the cases no other test covers.
func TestGakeLangRules(t *testing.T) {
	cases := []struct {
		name     string
		stub     string
		langFlag string
		wantLang utils.Language
		wantSrc  string
	}{
		{"new C, matching flag", "new-c", "c", utils.LangC, getCProjExpectedCmakeSrc("new-c", "core.c", "utils.c")},
		{"new C++, no flag", "new", "", utils.LangCXX, newProjExpectedCmakeSrc},
		{"new C++, matching flag", "new", "c++", utils.LangCXX, newProjExpectedCmakeSrc},
		{"no src, flag decides C++", "new-no-src", "c++", utils.LangCXX, getCXXProjExpectedCmakeSrc()},
		{"no src, no flag: defaults to C++", "new-no-src", "", utils.LangCXX, getCXXProjExpectedCmakeSrc()},
		{"existing C, matching flag", "existing-c", "c", utils.LangC, getCProjExpectedCmakeSrc("existing-c", "core.c", "utils.c")},
		{"existing C++, matching flag", "existing", "c++", utils.LangCXX, newProjExpectedCmakeSrc},
		{"existing C, no src: project() decides", "existing-c-no-src", "", utils.LangC, getCProjExpectedCmakeSrc("existing-c-no-src")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := createStubCopy(t, c.stub)

			if err := gake(root, getArgsWithLang(c.langFlag)); err != nil {
				t.Fatalf("unexpected err %v", err)
			}

			assertCmakeContent(t, utils.JoinPath(root, "src"), c.wantSrc)
			assertIncludeDirMatchesLanguage(t, root, c.wantLang)
		})
	}
}

// Every rejected combination fails with the right error and leaves the project untouched.
func TestGakeLangRulesErrors(t *testing.T) {
	cases := []struct {
		name     string
		stub     string
		langFlag string
		wantErr  error
	}{
		{"new C, contradicting flag", "new-c", "c++", utils.ErrLangMismatch},
		{"new C++, contradicting flag", "new", "c", utils.ErrLangMismatch},
		{"new mixed, c flag cannot bypass", "new-mixed", "c", utils.ErrMixedSources},
		{"new mixed, c++ flag cannot bypass", "new-mixed", "c++", utils.ErrMixedSources},
		{"existing mixed, no flag", "existing-mixed", "", utils.ErrMixedSources},
		{"existing mixed, matching flag cannot bypass", "existing-mixed", "c", utils.ErrMixedSources},
		{"existing C, sources are C++", "existing-c-cppm", "", utils.ErrLangMismatch},
		{"existing C, sources are C++, flag matches project()", "existing-c-cppm", "c", utils.ErrLangMismatch},
		{"existing C++, sources are C", "existing-cxx-c", "", utils.ErrLangMismatch},
		{"existing C++, sources are C, flag matches project()", "existing-cxx-c", "c++", utils.ErrLangMismatch},
		{"existing C, mismatching flag", "existing-c", "c++", utils.ErrLangMismatch},
		{"existing C++, mismatching flag", "existing", "c", utils.ErrLangMismatch},
		{"unknown flag value", "new-c", "rust", utils.ErrUnknownLang},
		{"unknown flag value, no src", "new-no-src", "rust", utils.ErrUnknownLang},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := createStubCopy(t, c.stub)
			before := snapshotDir(t, root)

			err := gake(root, getArgsWithLang(c.langFlag))
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("got err %v, want %v", err, c.wantErr)
			}

			assertDirUnchanged(t, root, before)
		})
	}
}

const langFlagExpectedCmakeTests = `function(AddTest TEST_NAME TEST_SOURCE)
    add_executable(${TEST_NAME}
        ${TEST_SOURCE}
    )

    target_link_libraries(${TEST_NAME}
        PRIVATE new-no-src
    )

    add_test(NAME ${TEST_NAME} COMMAND ${TEST_NAME})
endfunction()
AddTest(Core core.c)`

// A fresh project with no src/ becomes C only because of the flag.
func TestGakeLangFlag(t *testing.T) {
	root := createStubCopy(t, "new-no-src")

	if err := gake(root, getArgsWithLang("c")); err != nil {
		t.Fatal(err)
	}

	assertCmakeContent(t, root, cProjExpectedCmakeRoot)
	assertCmakeContent(t, utils.JoinPath(root, "src"), getCProjExpectedCmakeSrc("new-no-src"))
	assertCmakeContent(t, utils.JoinPath(root, "tests"), langFlagExpectedCmakeTests)
	assertDirExists(t, utils.JoinPath(root, "include"))
}

// --lang= with no value is an explicit, invalid choice, not the same as leaving the flag out.
func TestGakeEmptyLangValueFails(t *testing.T) {
	root := createStubCopy(t, "new-c")
	before := snapshotDir(t, root)

	err := gake(root, utils.Args{ProjName: testProjName(), Lang: "", IsLangSet: true})
	if !errors.Is(err, utils.ErrUnknownLang) {
		t.Fatalf("got err %v, want %v", err, utils.ErrUnknownLang)
	}

	assertDirUnchanged(t, root, before)
}
