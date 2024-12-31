package i18n

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const LocalizeTag = "i18n"

type Locale struct {
	lang        string
	translation Translation
}

func NewLocale(lang string, translation Translation) *Locale {
	return &Locale{
		lang:        lang,
		translation: translation,
	}
}

func (l Locale) Lang() string {
	return l.lang
}

func (l Locale) Get(key string) (string, error) {
	if !strings.ContainsAny(key, "{}") {
		key = "{" + key + "}"
	}
	return l.getWithProcess(key)
}

func (l Locale) getWithProcess(key string) (string, error) {
	processor := NewI18NKeyProcessor(key)
	values := []string{}
	errs := []error{}
	for _, k := range processor.GetKeys() {
		value, ok := l.translation[k]
		if !ok {
			err := NewLocalizationError("translation not found for %q", k)
			errs = append(errs, err)
			continue
		}
		values = append(values, value)
	}
	if len(errs) != 0 {
		return "", errors.Join(errs...)
	}
	processor.SetValues(values...)
	return processor.String()
}

func (l Locale) GetOr(key, defaultValue string) string {
	if value, err := l.Get(key); err == nil {
		return value
	}
	return defaultValue
}

type Setter func(string)

func (l Locale) LocalizeFields(setters map[string]Setter) error {
	errs := []error{}
	for key, setter := range setters {
		value, err := l.Get(key)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		setter(value)
	}

	return errors.Join(errs...)
}

func (l Locale) LocalizeStruct(obj any) error {
	ot := reflect.TypeOf(obj)
	if ot == nil || ot.Kind() != reflect.Struct {
		return NewLocalizationError("the value must be a struct")
	}

	ov := reflect.ValueOf(obj)

	errs := []error{}
	for i := 0; i < ov.NumField(); i++ {
		ft, fv := ot.Field(i), ov.Field(i)
		key, ok := ft.Tag.Lookup(LocalizeTag)
		if !ok {
			continue
		}

		if ft.Type.Kind() != reflect.String {
			errs = append(errs, NewLocalizationError("field %q must be a string", ft.Name))
			continue
		}

		if !fv.CanSet() {
			errs = append(errs, NewLocalizationError("cannot set value into field %q", ft.Name))
			continue
		}

		value, err := l.Get(key)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		fv.SetString(value)
	}

	return errors.Join(errs...)
}

type LocalizationError struct {
	message string
}

func NewLocalizationError(message string, args ...any) *LocalizationError {
	message = fmt.Sprintf(message, args...)
	return &LocalizationError{message: message}
}

func (e LocalizationError) Error() string {
	return fmt.Sprintf("localizor: %s", e.message)
}
