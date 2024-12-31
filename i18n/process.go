package i18n

import (
	"errors"
	"strings"
)

var ErrIvalidNumberOfValues = errors.New("invalid number of values")

// Can only split and concatinates values, for example:
// "and {key1} >= {key2}" => "and value1 >= value2"
type I18NKeyProcessor struct {
	template string
	values   []string
}

func NewI18NKeyProcessor(template string) *I18NKeyProcessor {
	return &I18NKeyProcessor{
		template: template,
		values:   []string{},
	}
}

func (p I18NKeyProcessor) GetKeys() []string {
	if strings.ContainsAny(p.template, "{}") {
		return process(p.template)
	}
	return []string{p.template}
}

func (p *I18NKeyProcessor) SetValues(values ...string) {
	p.values = values
}

func (p I18NKeyProcessor) String() (string, error) {
	keys := p.GetKeys()
	if len(keys) != len(p.values) {
		return "", ErrIvalidNumberOfValues
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
