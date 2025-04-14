package utils

import (
	"github.com/pemistahl/lingua-go"
)

var (
	detector  lingua.LanguageDetector
	languages = []lingua.Language{
		lingua.English,
		lingua.Russian,
		lingua.Ukrainian,
		lingua.French,
		lingua.German,
		lingua.Spanish,
		lingua.Portuguese,
		lingua.Italian,
		lingua.Japanese,
		lingua.Chinese,
		lingua.Hindi,
		lingua.Korean,
		lingua.Thai,
		lingua.Vietnamese,
		lingua.Hungarian,
	}
)

func DetectLanguage(text string) string {
	if detector == nil {
		detector = lingua.NewLanguageDetectorBuilder().
			FromLanguages(languages...).
			Build()
	}

	language, exists := detector.DetectLanguageOf(text)
	if !exists {
		return "unknown"
	}

	return language.String()
}
