package validation

import (
	"fmt"
	"strings"
)

// ValidationError represetns an error encountered furing validation.
// It includes a descriptive message and optionally the name of the field
// which is validated.
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

// GetI18NKey generates an internationalization key for
// the validation error. The returned string includes the field name
// (if present) and the message, both converted to kebab-case and
// wrapped in curly braces, separated by space.
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
