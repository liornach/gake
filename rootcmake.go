package main

import (
	"embed"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	rootCmakeTmpl = "CMakeLists.txt.tmpl"
)

//go:embed CMakeLists.txt.tmpl
var rootEmbed embed.FS

type cmakeRootContent struct {
	core        string
	userContent string
}

func parseRootCmake(cmakeFilePath string) (cmakeRootContent, error) {
	projName, err := getProjNameFromCmake(cmakeFilePath)
	if err != nil {
		return cmakeRootContent{}, err
	}

	fullCmake, err := os.ReadFile(cmakeFilePath)
	if err != nil {
		return cmakeRootContent{}, err
	}

	core, err := insertProjnameToTemplate(projName)
	if err != nil {
		return cmakeRootContent{}, err
	}

	userContent := strings.TrimPrefix(string(fullCmake), core)
	return cmakeRootContent{
		core:        core,
		userContent: userContent,
	}, nil
}

func insertProjnameToTemplate(projname string) (string, error) {
	rendered, err := insertDataToTemplate(rootEmbed, rootCmakeTmpl, cmakeProjname{ProjName: projname})
	if err != nil {
		return "", err
	}

	return string(rendered), nil
}

func getProjNameFromCmake(root string) (string, error) {
	data, err := os.ReadFile(joinCmakeLists(root))
	if err != nil {
		return "", err
	}

	// case-insensitive match for project(...)
	re := regexp.MustCompile(`(?i)project\s*\(\s*([^\s\)]+)`)

	m := re.FindSubmatch(data)
	if m == nil {
		return "", fmt.Errorf("project() not found")
	}

	return string(m[1]), nil
}
