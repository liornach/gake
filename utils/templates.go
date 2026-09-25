package utils

import (
	"bytes"
	"embed"
	"os"
	"text/template"
)

//go:embed *.tmpl
var templatesFS embed.FS

type cmakeProjname struct {
	ProjName string
}

func insertDataToTemplate(tmplFileName string, data any) ([]byte, error) {
	tmplBytes, err := templatesFS.ReadFile(tmplFileName)
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

type srcTemplateData struct {
	LibName string
	Sources []string
}

func readExistingCmake(root string) ([]byte, error) {
	return os.ReadFile(JoinCmakeLists(root))
}

type testsTemplateData struct {
	LibName string
	Tests   []TestEntry
}

func RenderRootCmake(profile Profile, projName string) ([]byte, error) {
	return insertDataToTemplate(profile.rootTmpl, cmakeProjname{ProjName: projName})
}

func RenderSrcCmake(profile Profile, sources []string) ([]byte, error) {
	return insertDataToTemplate(profile.srcTmpl, srcTemplateData{LibName: profile.LibName, Sources: sources})
}

func RenderTestsCmake(profile Profile, tests []TestEntry) ([]byte, error) {
	return insertDataToTemplate(testsCmakeTmpl, testsTemplateData{LibName: profile.LibName, Tests: tests})
}
