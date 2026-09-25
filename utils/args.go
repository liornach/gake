package utils

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

var ErrMissingProjName = errors.New("a new project needs a name: pass --projname=<projname>")

type Args struct {
	ProjName  string
	Lang      string
	IsLangSet bool // true when --lang was passed, even with an empty value
}

// newFlagSet declares gake's flags, storing their values into parsed.
func newFlagSet(parsed *Args, out io.Writer) *flag.FlagSet {
	flags := flag.NewFlagSet("gake", flag.ContinueOnError)
	flags.SetOutput(out)
	flags.StringVar(&parsed.ProjName, "projname", "", "the name of the cmake project (new projects only)")
	flags.StringVar(&parsed.Lang, "lang", "", "project language: c or c++ (default: detected from src/, else c++)")
	return flags
}

// ParseArgs parses gake's command line (without the program name) into Args.
// Values are returned raw; validating them is the caller's job. -h returns flag.ErrHelp.
func ParseArgs(args []string) (Args, error) {
	var parsed Args
	flags := newFlagSet(&parsed, io.Discard)
	if err := flags.Parse(args); err != nil {
		return Args{}, err
	}

	parsed.IsLangSet = isFlagSet(flags, "lang")
	return parsed, nil
}

// WriteUsage writes gake's usage and flag list to out.
func WriteUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage: gake [--projname=<name>] [--lang=c|c++]")
	newFlagSet(&Args{}, out).PrintDefaults()
}

func isFlagSet(flags *flag.FlagSet, name string) bool {
	isSet := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == name {
			isSet = true
		}
	})

	return isSet
}
