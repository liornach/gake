package utils

import (
	"errors"
	"testing"
)

func TestParseArgs(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want Args
	}{
		{"no flags", nil, Args{}},
		{"projname only", []string{"--projname=x"}, Args{ProjName: "x"}},
		{"lang only", []string{"--lang=c"}, Args{Lang: "c", IsLangSet: true}},
		{"both", []string{"--projname=x", "--lang=c++"}, Args{ProjName: "x", Lang: "c++", IsLangSet: true}},
		{"lang is not validated here", []string{"--lang=rust"}, Args{Lang: "rust", IsLangSet: true}},
		{"empty lang value is still set", []string{"--lang="}, Args{Lang: "", IsLangSet: true}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseArgs(c.args)
			if err != nil {
				t.Fatal(err)
			}

			if got != c.want {
				t.Fatalf("ParseArgs(%q) = %+v, want %+v", c.args, got, c.want)
			}
		})
	}
}

func TestParseArgsUnknownFlag(t *testing.T) {
	if _, err := ParseArgs([]string{"--bogus=1"}); err == nil {
		t.Fatal("expected an error for an unknown flag")
	}
}

func TestGetLanguageFromFlag(t *testing.T) {
	cases := []struct {
		value string
		want  Language
	}{
		{"c", LangC},
		{"c++", LangCXX},
	}

	for _, c := range cases {
		t.Run(c.value, func(t *testing.T) {
			got, err := GetLanguageFromFlag(c.value)
			if err != nil {
				t.Fatal(err)
			}

			if got != c.want {
				t.Fatalf("GetLanguageFromFlag(%q) = %v, want %v", c.value, got, c.want)
			}
		})
	}
}

func TestGetLanguageFromFlagUnknown(t *testing.T) {
	for _, value := range []string{"C", "cxx", "rust", ""} {
		t.Run(value, func(t *testing.T) {
			_, err := GetLanguageFromFlag(value)
			if !errors.Is(err, ErrUnknownLang) {
				t.Fatalf("GetLanguageFromFlag(%q): got err %v, want %v", value, err, ErrUnknownLang)
			}
		})
	}
}
