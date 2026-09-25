package utils

import (
	"fmt"
	"slices"
)

const (
	cxxRootCmakeTmpl = "CMakeLists.txt.tmpl"
	testsCmakeTmpl   = "TestsCMakeLists.txt.tmpl"

	cxxLibName = "Objects"
)

// languageSpec holds the fixed facts about one language. Everything that differs
// between C and C++ lives here, so no other code needs to branch on the language.
type languageSpec struct {
	lang                Language
	flagName            string // value of --lang
	cmakeName           string // language name in project()
	srcExt              string
	testExt             string
	rootTmpl            string
	srcTmpl             string
	extraDirs           []string // created in the project root
	isLibNamedAfterRoot bool     // otherwise the library is cxxLibName
}

var languageSpecs = []languageSpec{
	{
		lang:      LangCXX,
		flagName:  "c++",
		cmakeName: "CXX",
		srcExt:    ".cppm",
		testExt:   ".cpp",
		rootTmpl:  cxxRootCmakeTmpl,
		srcTmpl:   "SrcCMakeLists.txt.tmpl",
	},
	{
		lang:                LangC,
		flagName:            "c",
		cmakeName:           "C",
		srcExt:              ".c",
		testExt:             ".c",
		rootTmpl:            "CCMakeLists.txt.tmpl",
		srcTmpl:             "CSrcCMakeLists.txt.tmpl",
		extraDirs:           []string{"include"},
		isLibNamedAfterRoot: true,
	},
}

func getLanguageSpec(lang Language) (languageSpec, bool) {
	for _, spec := range languageSpecs {
		if spec.lang == lang {
			return spec, true
		}
	}

	return languageSpec{}, false
}

// Profile is a language spec applied to one project root.
type Profile struct {
	SrcExt    string
	TestExt   string
	LibName   string
	ExtraDirs []string
	rootTmpl  string
	srcTmpl   string
}

// GetProfile returns the profile for lang in root. It fails for LangUndecided
// (the caller must apply a default first) and for an invalid C library name.
func GetProfile(lang Language, root string) (Profile, error) {
	spec, isKnown := getLanguageSpec(lang)
	if !isKnown {
		return Profile{}, fmt.Errorf("no profile for language %v", lang)
	}

	libName, err := getProfileLibName(spec, root)
	if err != nil {
		return Profile{}, err
	}

	return Profile{
		SrcExt:    spec.srcExt,
		TestExt:   spec.testExt,
		LibName:   libName,
		ExtraDirs: slices.Clone(spec.extraDirs),
		rootTmpl:  spec.rootTmpl,
		srcTmpl:   spec.srcTmpl,
	}, nil
}

func getProfileLibName(spec languageSpec, root string) (string, error) {
	if spec.isLibNamedAfterRoot {
		return GetLibName(root)
	}

	return cxxLibName, nil
}
