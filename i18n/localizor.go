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

type Localizor struct {
	translations map[string]Translation
}

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

func (l *Localizor) GetLocale(langs ...string) (Locale, error) {
	for _, lang := range langs {
		if translation, ok := l.translations[lang]; ok {
			return *NewLocale(lang, translation), nil
		}
	}
	return Locale{}, ErrLanguageNotFound
}

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
	lastIndex := strings.LastIndex(filename, path.Ext(filename))
	return filename[:lastIndex]
}
