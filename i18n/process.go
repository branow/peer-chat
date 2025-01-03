package i18n

import (
	"errors"
	"strings"
)

var ErrInvalidNumberOfValues = errors.New("invalid number of values")

// I18NKeyProcessor is responsible for processing a template string
// and replacing placeholders (e.g. {key}) with corresponding values.
type I18NKeyProcessor struct {
	template string
	values   []string
}

// NewI18NKeyProcessor creates a new processor for a given template.
func NewI18NKeyProcessor(template string) *I18NKeyProcessor {
	return &I18NKeyProcessor{
		template: template,
		values:   []string{},
	}
}

// GetKeys extracts the keys (placeholders) from the template string.
// It searches for substrings enclosed in curly braces, such as {key1}.
// If no placeholders are found, it returns the template as a stingle key.
func (p I18NKeyProcessor) GetKeys() []string {
	if strings.ContainsAny(p.template, "{}") {
		return process(p.template)
	}
	return []string{p.template}
}

// SetValues sets the values that will replace the placeholders in the template.
func (p *I18NKeyProcessor) SetValues(values ...string) {
	p.values = values
}

// String processes the template by replacing the keys with the corresponding values.
// It returns an error if the number of values does not mathc the number of keys.
func (p I18NKeyProcessor) String() (string, error) {
	keys := p.GetKeys()
	if len(keys) != len(p.values) {
		return "", ErrInvalidNumberOfValues
	}
	keysValues := map[string]string{}
	for i, key := range keys {
		keysValues[key] = p.values[i]
	}
	return replace(p.template, keysValues), nil
}

func process(template string) []string {
	keys := []string{}
	chars := []rune(template)
	in := false
	start := 0
	for i, char := range chars {
		if char == '{' && !in {
			start = i + 1
			in = true
		} else if char == '}' && in {
			keys = append(keys, template[start:i])
			in = false
		}
	}
	return keys
}

func replace(template string, keys map[string]string) string {
	result := template
	for key, value := range keys {
		result = strings.Replace(result, "{"+key+"}", value, 1)
	}
	return result
}
