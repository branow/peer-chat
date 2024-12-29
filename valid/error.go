package valid

type ValidationError struct {
	Message string
	Field   string
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return e.Field + " " + e.Message
}
