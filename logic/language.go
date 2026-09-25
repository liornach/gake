package logic

import (
	"fmt"
	"gake/utils"
)

// resolveNewProjLanguage decides a new project's language from src/ and the --lang flag.
func resolveNewProjLanguage(root string, args utils.Args) (utils.Language, error) {
	flagLang, err := getFlagLanguage(args)
	if err != nil {
		return utils.LangUndecided, err
	}

	srcLang, err := utils.DetectLanguageFromSources(utils.GetSrcDir(root))
	if err != nil {
		return utils.LangUndecided, err
	}

	lang, err := reconcileLanguages(srcLang, flagLang, "--lang")
	if err != nil {
		return utils.LangUndecided, err
	}

	return getLanguageOrDefault(lang), nil
}

// resolveExistingProjLanguage takes the language from project(). src/ and --lang may
// only agree with it (or, for src/, say nothing).
func resolveExistingProjLanguage(root string, args utils.Args) (utils.Language, error) {
	flagLang, err := getFlagLanguage(args)
	if err != nil {
		return utils.LangUndecided, err
	}

	cmakeLang, err := utils.ReadCmakeLanguage(root)
	if err != nil {
		return utils.LangUndecided, err
	}

	srcLang, err := utils.DetectLanguageFromSources(utils.GetSrcDir(root))
	if err != nil {
		return utils.LangUndecided, err
	}

	if _, err := reconcileLanguages(cmakeLang, srcLang, "src/"); err != nil {
		return utils.LangUndecided, err
	}

	return reconcileLanguages(cmakeLang, flagLang, "--lang")
}

// getFlagLanguage returns the --lang language, or LangUndecided when the flag wasn't passed.
func getFlagLanguage(args utils.Args) (utils.Language, error) {
	if !args.IsLangSet {
		return utils.LangUndecided, nil
	}

	return utils.GetLanguageFromFlag(args.Lang)
}

// reconcileLanguages merges two language sources: an undecided one defers to the other,
// two decided ones must be equal.
func reconcileLanguages(base, other utils.Language, otherName string) (utils.Language, error) {
	switch {
	case other == utils.LangUndecided:
		return base, nil
	case base == utils.LangUndecided:
		return other, nil
	case base != other:
		return utils.LangUndecided, fmt.Errorf("%w: project is %v but %s says %v", utils.ErrLangMismatch, base, otherName, other)
	default:
		return base, nil
	}
}

// getLanguageOrDefault applies the C++ default when nothing decided the language.
func getLanguageOrDefault(lang utils.Language) utils.Language {
	if lang == utils.LangUndecided {
		return utils.LangCXX
	}

	return lang
}
