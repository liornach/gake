package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
)

func projnameCmake(cmakeListsPath string) (string, error) {
	data, err := os.ReadFile(cmakeListsPath)
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

func projnameArgv() string {
	var projName string
	flag.StringVar(&projName, "projname", "", "the name of the cmake project")
	flag.Parse()
	return projName
}

func mkdir(fullpath string) error {
	return os.MkdirAll(fullpath, 0755)
}

func initializeCmake(root string) error {
	srcDir := root + "/src"
	if err := mkdir(srcDir); err != nil {
		return err
	}

	testsDir := root + "/tests"
	if err := mkdir(testsDir); err != nil {
		return err
	}

	return nil
}

func containCmakeLists(path string) (bool, error) {
	fullPath := path + "/CMakeLists.txt"
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}

	return err == nil, err
}

func updateCmake(root string) error {
	contain, err := containCmakeLists(root)
	if err != nil {
		return err
	}

	if !contain { // looking good, need to polish a little since this is not the final logic, initializeCmake and updateCmake probably are NOT both needed
		initializeCmake(root)
	}

	return nil
}

func wd() (string, error) {
	return os.Getwd()
}

func main() {

	updateCmake() // call this fucntion with the directory from WHERE this utility was called, not from
	// where it resides

	fmt.Println("Hello, World!")
}
