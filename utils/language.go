package utils

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

type Language int

const (
	LangUndecided Language = iota
	LangCXX
	LangC
)

func (l Language) String() string {
	spec, isKnown := getLanguageSpec(l)
	if !isKnown {
		return "undecided"
	}

	return spec.cmakeName
}

var (
	ErrMixedSources = errors.New("mixed C and C++ module sources in src/, cannot infer language")
	ErrLangMismatch = errors.New("language mismatch")
	ErrUnknownLang  = errors.New("unknown --lang value (allowed: c, c++)")
)

// getProjectLanguages returns the known languages listed in project(), in any case.
func getProjectLanguages(content []byte) []Language {
	var languages []Language
	for _, arg := range getProjectLanguageArgs(getProjectArgs(content)) {
		for _, spec := range languageSpecs {
			if strings.EqualFold(arg, spec.cmakeName) {
				languages = append(languages, spec.lang)
			}
		}
	}

	return languages
}

func isCOnlyProject(languages []Language) bool {
	return slices.Contains(languages, LangC) && !slices.Contains(languages, LangCXX)
}

// getLanguageFromCmake reads the language of a root CMakeLists.txt from its project() call.
// CXX wins over C; no language listed means CXX, as before C support existed.
func getLanguageFromCmake(content []byte) Language {
	if isCOnlyProject(getProjectLanguages(content)) {
		return LangC
	}

	return LangCXX
}

// ReadCmakeLanguage reads root/CMakeLists.txt and returns the language its project() declares.
func ReadCmakeLanguage(root string) (Language, error) {
	content, err := readExistingCmake(root)
	if err != nil {
		return LangUndecided, err
	}

	return getLanguageFromCmake(content), nil
}

func getLanguageBySrcExt(ext string) (Language, bool) {
	for _, spec := range languageSpecs {
		if spec.srcExt == ext {
			return spec.lang, true
		}
	}

	return LangUndecided, false
}

// DetectLanguageFromSources decides the language from the source files in srcDir.
// Headers don't count, and an empty or missing srcDir is LangUndecided.
func DetectLanguageFromSources(srcDir string) (Language, error) {
	entries, err := readDirIfExists(srcDir)
	if err != nil {
		return LangUndecided, err
	}

	found := map[Language]bool{}
	for _, entry := range entries {
		lang, isSource := getLanguageBySrcExt(filepath.Ext(entry.Name()))
		if isSource && !entry.IsDir() {
			found[lang] = true
		}
	}

	return getLanguageFromFound(found)
}

func getLanguageFromFound(found map[Language]bool) (Language, error) {
	switch len(found) {
	case 0:
		return LangUndecided, nil
	case 1:
		return getOnlyKey(found), nil
	default:
		return LangUndecided, ErrMixedSources
	}
}

func getOnlyKey(found map[Language]bool) Language {
	for lang := range found {
		return lang
	}

	return LangUndecided
}

// GetLanguageFromFlag maps a --lang value to a Language. Only "c" and "c++" are valid.
func GetLanguageFromFlag(value string) (Language, error) {
	for _, spec := range languageSpecs {
		if spec.flagName == value {
			return spec.lang, nil
		}
	}

	return LangUndecided, fmt.Errorf("%w: %q", ErrUnknownLang, value)
}
