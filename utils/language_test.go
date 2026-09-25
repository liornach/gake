package utils

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestGetLanguageFromCmake(t *testing.T) {
	cases := []struct {
		name  string
		cmake string
		want  Language
	}{
		{"only C", "project(x C)", LangC},
		{"only CXX", "project(x CXX)", LangCXX},
		{"C and CXX picks CXX", "project(x C CXX)", LangCXX},
		{"LANGUAGES keyword", "project(x VERSION 1.2 LANGUAGES C)", LangC},
		{"no languages defaults to CXX", "project(x)", LangCXX},
		{"version only defaults to CXX", "project(x VERSION 1.2)", LangCXX},
		{"lowercase language", "project(x c)", LangC},
		{"uppercase command", "PROJECT(x C)", LangC},
		{"multiline", "project(\n    x\n    LANGUAGES C\n)", LangC},
		{"full root file", "cmake_minimum_required(VERSION 3.28)\nproject(vm C)\n\nset(CMAKE_C_STANDARD 23)", LangC},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := getLanguageFromCmake([]byte(c.cmake))
			if got != c.want {
				t.Fatalf("getLanguageFromCmake(%q) = %v, want %v", c.cmake, got, c.want)
			}
		})
	}
}

func createEmptyFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDetectLanguageFromSources(t *testing.T) {
	cases := []struct {
		name  string
		files []string
		want  Language
	}{
		{"only .c files", []string{"a.c", "b.c"}, LangC},
		{"only .cppm files", []string{"a.cppm", "b.cppm"}, LangCXX},
		{"empty dir is undecided", nil, LangUndecided},
		{"headers alone are undecided", []string{"a.h"}, LangUndecided},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srcDir := t.TempDir()
			createEmptyFiles(t, srcDir, c.files...)

			got, err := DetectLanguageFromSources(srcDir)
			if err != nil {
				t.Fatal(err)
			}

			if got != c.want {
				t.Fatalf("DetectLanguageFromSources(%v) = %v, want %v", c.files, got, c.want)
			}
		})
	}
}

func TestDetectLanguageFromSourcesMissingDir(t *testing.T) {
	missingDir := filepath.Join(t.TempDir(), "src")

	got, err := DetectLanguageFromSources(missingDir)
	if err != nil {
		t.Fatal(err)
	}

	if got != LangUndecided {
		t.Fatalf("missing src dir: got %v, want %v", got, LangUndecided)
	}
}

func TestDetectLanguageFromSourcesMixed(t *testing.T) {
	srcDir := t.TempDir()
	createEmptyFiles(t, srcDir, "a.c", "b.cppm")

	_, err := DetectLanguageFromSources(srcDir)
	if !errors.Is(err, ErrMixedSources) {
		t.Fatalf("got err %v, want %v", err, ErrMixedSources)
	}
}

func TestGetLibName(t *testing.T) {
	cases := []struct {
		dirName string
		isValid bool
	}{
		{"calc", true},
		{"new-c", true},
		{"under_score", true},
		{"Lib2", true},
		{"lib.v2", false},
		{"a+b", false},
		{"my lib", false},
		{"bad$name", false},
		{"dir(1)", false},
	}

	for _, c := range cases {
		t.Run(c.dirName, func(t *testing.T) {
			root := filepath.Join("/home/me/projects", c.dirName)
			got, err := GetLibName(root)

			if !c.isValid {
				if !errors.Is(err, ErrInvalidLibName) {
					t.Fatalf("GetLibName(%q): got err %v, want %v", root, err, ErrInvalidLibName)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetLibName(%q): unexpected err %v", root, err)
			}
			if got != c.dirName {
				t.Fatalf("GetLibName(%q) = %q, want %q", root, got, c.dirName)
			}
		})
	}
}
