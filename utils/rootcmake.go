package utils

import (
	"fmt"
	"os"
	"strings"
)

type cmakeRootContent struct {
	core        string
	userContent string
}

func parseRootCmake(cmakeFilePath string) (cmakeRootContent, error) {
	projName, err := readProjNameFromCmake(cmakeFilePath)
	if err != nil {
		return cmakeRootContent{}, err
	}

	fullCmake, err := os.ReadFile(cmakeFilePath)
	if err != nil {
		return cmakeRootContent{}, err
	}

	core, err := InsertProjnameToTemplate(projName)
	if err != nil {
		return cmakeRootContent{}, err
	}

	userContent := strings.TrimPrefix(string(fullCmake), core)
	return cmakeRootContent{
		core:        core,
		userContent: userContent,
	}, nil
}

func InsertProjnameToTemplate(projname string) (string, error) {
	rendered, err := insertDataToTemplate(cxxRootCmakeTmpl, cmakeProjname{ProjName: projname})
	if err != nil {
		return "", err
	}

	return string(rendered), nil
}

func readProjNameFromCmake(root string) (string, error) {
	data, err := readExistingCmake(root)
	if err != nil {
		return "", err
	}

	args := getProjectArgs(data)
	if len(args) == 0 {
		return "", fmt.Errorf("project() not found")
	}

	return args[0], nil
}
