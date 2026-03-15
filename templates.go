package main

import (
	"bytes"
	"embed"
	"os"
	"text/template"
)

//go:embed SrcCMakeLists.txt.tmpl
var srcCmakeEmbed embed.FS

//go:embed TestsCMakeLists.txt.tmpl
var testsCmakeEmbed embed.FS

const (
	srcCmakeTmpl   = "SrcCMakeLists.txt.tmpl"
	testsCmakeTmpl = "TestsCMakeLists.txt.tmpl"
)

type cmakeProjname struct {
	ProjName string
}

func insertDataToTemplate(embedFs embed.FS, tmplFileName string, data any) ([]byte, error) {
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

func insertDataToSrcTemplate(cppmFiles []string) ([]byte, error) {
	return insertDataToTemplate(srcCmakeEmbed, srcCmakeTmpl, cmakeCppmFiles{CppmFiles: cppmFiles})
}

func readExistingCmake(root string) ([]byte, error) {
	return os.ReadFile(joinCmakeLists(root))
}

type TestsTemplateData struct {
	Tests []TestEntry
}

func insertDataToTestsTemplate(tests []TestEntry) ([]byte, error) {
	return insertDataToTemplate(testsCmakeEmbed, testsCmakeTmpl, TestsTemplateData{
		Tests: tests,
	})
}
