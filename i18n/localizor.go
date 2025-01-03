package i18n

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"strings"
)

var (
	ErrLanguageNotFound = errors.New("localizor: language not found")
)

type Translation map[string]string

// Localizor handles loading and managing translation
// for multiple languages.
type Localizor struct {
	translations map[string]Translation
}

// NewLocalizor initializes a new Localizor instance
// by reading translation files from the specified directory.
// It processes all files it can, if any files cannot be processed
// (e.g., due to errors in reading or parsing), those errors are
// collected and returned.
// File names must follow the pattern 'lang.json'
// (for example: en.json, ua.json).
func NewLocalizor(dir string) (*Localizor, error) {
	localizor := Localizor{map[string]Translation{}}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return &localizor, err
	}

	errs := []error{}
	for _, e := range entries {
		filepath := path.Join(dir, e.Name())
		if err := localizor.consumeFile(filepath); err != nil {
			errs = append(errs, err)
		}
	}

	return &localizor, errors.Join(errs...)
}

// GetLocale returns the locale for the first found language
// or an error if none are found.
func (l *Localizor) GetLocale(langs ...string) (Locale, error) {
	for _, lang := range langs {
		if translation, ok := l.translations[lang]; ok {
			return *NewLocale(lang, translation), nil
		}
	}
	return Locale{}, ErrLanguageNotFound
}

// HasLanguage checks if a translation for a given langauge exists.
func (l *Localizor) HasLanguage(lang string) bool {
	_, ok := l.translations[lang]
	return ok
}

func (l *Localizor) consumeFile(filepath string) error {
	translation, err := readTranslationFromFile(filepath)
	if err != nil {
		return err
	}
	lang := extractFilename(filepath)
	l.addTranslation(lang, translation)
	return nil
}

func (l *Localizor) addTranslation(lang string, translation Translation) {
	l.translations[lang] = translation
}

func readTranslationFromFile(filepath string) (Translation, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	return readTranslation(file)
}

func readTranslation(r io.Reader) (Translation, error) {
	translation := Translation{}
	if err := json.NewDecoder(r).Decode(&translation); err != nil {
		return nil, err
	}
	return translation, nil
}

func extractFilename(filepath string) string {
	filename := path.Base(filepath)
	return strings.TrimSuffix(filename, path.Ext(filename))
}
