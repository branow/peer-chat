package valid

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Message string
	Field   string
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

func (e ValidationError) Error() string {
	message := e.Message
	if e.Field != "" {
		message = e.Field + " " + message
	}
	return message
}

func (e ValidationError) GetI18NKey() string {
	key := toI18NKey(e.Message)
	if e.Field != "" {
		return fmt.Sprintf("{%s} {%s}", toI18NKey(e.Field), key)
	}
	return key
}

func toI18NKey(str string) string {
	str = strings.TrimSpace(str)
	str = strings.ToLower(str)
	str = strings.ReplaceAll(str, " ", "-")
	return str
}
