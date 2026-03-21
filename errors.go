package main

import (
	"fmt"
	"runtime"
)

type noProjNameErr struct {
}

func (noProjNameErr) Error() string {
	return "project name could not be disolved. Make sure that either :\n-you have a CMakeLists.txt file containing project() statement\n-you provided this utility with --projname=<projname> argument"
}

type notMadeByUsError struct {
}

func (notMadeByUsError) Error() string {
	return "Looks like the existing CMakeLists.txt file was not made by us, since it is not matching our expected template"
}

type missingDirError struct {
	dir string
}

func (e missingDirError) Error() string {
	return fmt.Sprintf("expected directory structure missing: %s", e.dir)
}

func callerFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func wrapError(msg string, err error) error {
	return fmt.Errorf("%s, err : %w", msg, err)
}
