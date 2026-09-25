package utils

import (
	"slices"
	"strings"
)

// cmakeToken is one CMake token: a word, a quoted argument, "(" or ")".
type cmakeToken struct {
	text     string
	isQuoted bool
}

func (t cmakeToken) isParen(paren string) bool {
	return !t.isQuoted && t.text == paren
}

// tokenizeCmake splits CMake source into tokens, dropping # comments.
// A quoted argument is one token, even when it holds spaces, parentheses or #.
func tokenizeCmake(text string) []cmakeToken {
	var tokens []cmakeToken
	var word strings.Builder

	flushWord := func() {
		if word.Len() > 0 {
			tokens = append(tokens, cmakeToken{text: word.String()})
			word.Reset()
		}
	}

	for i := 0; i < len(text); i++ {
		switch ch := text[i]; {
		case ch == '#':
			flushWord()
			i = getLineEnd(text, i)
		case ch == '"':
			flushWord()
			var quoted string
			quoted, i = readQuoted(text, i)
			tokens = append(tokens, cmakeToken{text: quoted, isQuoted: true})
		case ch == '(' || ch == ')':
			flushWord()
			tokens = append(tokens, cmakeToken{text: string(ch)})
		case isCmakeSpace(ch):
			flushWord()
		default:
			word.WriteByte(ch)
		}
	}

	flushWord()
	return tokens
}

func isCmakeSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// getLineEnd returns the index of the newline ending the line that contains i.
func getLineEnd(text string, i int) int {
	end := strings.IndexByte(text[i:], '\n')
	if end < 0 {
		return len(text)
	}

	return i + end
}

// readQuoted reads the quoted argument opening at text[start] and returns its
// content and the index of the closing quote. Backslash escapes are kept as-is.
func readQuoted(text string, start int) (string, int) {
	i := start + 1
	for i < len(text) && text[i] != '"' {
		if text[i] == '\\' {
			i++
		}
		i++
	}

	return text[start+1 : min(i, len(text))], i
}

// getProjectArgs returns the arguments of the first real project() call, name first.
func getProjectArgs(content []byte) []string {
	tokens := tokenizeCmake(string(content))
	for i := 0; i+1 < len(tokens); i++ {
		if isProjectCallAt(tokens, i) {
			return getCallArgs(tokens[i+2:])
		}
	}

	return nil
}

func isProjectCallAt(tokens []cmakeToken, i int) bool {
	return !tokens[i].isQuoted && strings.EqualFold(tokens[i].text, "project") && tokens[i+1].isParen("(")
}

// getCallArgs collects argument tokens up to the ")" closing the call.
func getCallArgs(tokens []cmakeToken) []string {
	var args []string
	depth := 1
	for _, token := range tokens {
		switch {
		case token.isParen("("):
			depth++
		case token.isParen(")"):
			depth--
			if depth == 0 {
				return args
			}
		default:
			args = append(args, token.text)
		}
	}

	return args
}

// project() keywords that take exactly one value, which is never a language.
var projectValueKeywords = []string{"VERSION", "DESCRIPTION", "HOMEPAGE_URL"}

// getProjectLanguageArgs returns the arguments of project() that name languages:
// the bare ones after the project name and the ones after LANGUAGES.
func getProjectLanguageArgs(args []string) []string {
	var languages []string
	for i := 1; i < len(args); i++ {
		keyword := strings.ToUpper(args[i])
		switch {
		case isProjectValueKeyword(keyword):
			i++
		case keyword == "LANGUAGES":
		default:
			languages = append(languages, args[i])
		}
	}

	return languages
}

func isProjectValueKeyword(keyword string) bool {
	return slices.Contains(projectValueKeywords, keyword)
}
