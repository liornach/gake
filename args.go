package main

import "flag"

func projnameArgv() (string, error) {
	var projName string
	flag.StringVar(&projName, "projname", "", "the name of the cmake project")
	flag.Parse()
	if projName == "" {
		return "", noProjNameErr{}
	}

	return projName, nil
}
