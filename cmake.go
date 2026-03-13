package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"regexp"
	"text/template"
)

//go:embed CMakeLists.txt.tmpl
var rootEmbed embed.FS

//go:embed SrcCMakeLists.txt.tmpl
var srcCmakeEmbed embed.FS

//go:embed TestsCMakeLists.txt.tmpl
var testsCmakeEmbed embed.FS

const rootCmakeTmpl = "CMakeLists.txt.tmpl"
const srcCmakeTmpl = "SrcCMakeLists.txt.tmpl"
const testsCmakeTmpl = "TestsCMakeLists.txt.tmpl"

type cmakeProjname struct {
	ProjName string
}

func renderRootCmakeTmpl(projname string) ([]byte, error) {
	return renderTemplate(rootEmbed, rootCmakeTmpl, cmakeProjname{ProjName: projname})
}

func renderTemplate(embedFs embed.FS, tmplFileName string, data any) ([]byte, error) {
	tmplBytes, err := embedFs.ReadFile(tmplFileName)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(tmplFileName).Parse(string(tmplBytes))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type cmakeCppmFiles struct {
	CppmFiles []string
}

func renderSrcCmakeTmpl(cppmFiles []string) ([]byte, error) {
	return renderTemplate(srcCmakeEmbed, srcCmakeTmpl, cmakeCppmFiles{CppmFiles: cppmFiles})
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

func readExistingCmake(root string) ([]byte, error) {
	return os.ReadFile(joinCmakeLists(root))
}

type TestsTemplateData struct {
	Tests []TestEntry
}

func renderTestsCmakeTmpl(tests []TestEntry) ([]byte, error) {
	return renderTemplate(testsCmakeEmbed, testsCmakeTmpl, TestsTemplateData{
		Tests: tests,
	})
}
