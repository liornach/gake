package utils

import "flag"

type noProjNameErr struct {
}

func (noProjNameErr) Error() string {
	return "project name could not be disolved. Make sure that either :\n-you have a CMakeLists.txt file containing project() statement\n-you provided this utility with --projname=<projname> argument"
}

func ProjnameArgv() (string, error) {
	var projName string
	flag.StringVar(&projName, "projname", "", "the name of the cmake project")
	flag.Parse()
	if projName == "" {
		return "", noProjNameErr{}
	}

	return projName, nil
}
