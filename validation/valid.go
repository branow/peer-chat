package validation

import (
	"fmt"
	"strconv"
	"strings"
)

// Validate performs a series of validation checks on the provided object
// based on the given constraints. If any constraint fails, the corresponding
// error is returned with the field name attached.
func Validate[T any](obj T, fieldName string, constraints ...Constraint[T]) error {
	for _, constraint := range constraints {
		if err := constraint(obj); err != nil {
			err.Field = fieldName
			return err
		}
	}
	return nil
}

// Constraint represents a validation functions that takes an object and
// retuns an error if validation fails.
type Constraint[T any] func(T) *ValidationError

// Equal checks if the value of obj is equal to any of the values in
// the provided list.
func Equal[T comparable](values []T, messages ...string) Constraint[T] {
	check := func(t T) bool {
		for _, value := range values {
			if t == value {
				return false
			}
		}
		return true
	}
	return makeConstraint(check, fmt.Sprintf("is not %v", values), messages...)
}

// NotEmpty checks if the string is not empty.
// messages are optional parameter to replace the default
// error message.
func NotEmpty(messages ...string) Constraint[string] {
	check := func(s string) bool { return len(s) == 0 }
	return makeConstraint(check, "is mandatory", messages...)
}

// NotEmpty checks if the string is not blank.
// messages are optional parameter to replace the default
// error message.
func NotBlank(message ...string) Constraint[string] {
	check := func(s string) bool { return len(strings.TrimSpace(s)) == 0 }
	return makeConstraint(check, "is mandatory", message...)
}

// NotLongerThan checks if the string is not longer than the specified value.
// messages are optional parameter to replace the default
// error message.
func NotLongerThan(value int, messages ...string) Constraint[string] {
	check := func(s string) bool { return len(s) > value }
	return makeConstraint(check, "is too long", messages...)
}

// NotShorterThan checks if the string is not shorter than the specified value.
// messages are optional parameter to replace the default
// error message.
func NotShorterThan(value int, messages ...string) Constraint[string] {
	check := func(s string) bool { return len(s) < value }
	return makeConstraint(check, "is too short", messages...)
}

// AnInteger checks if the string is a valid integer.
// messages are optional parameter to replace the default
// error message.
func AnInteger(message ...string) Constraint[string] {
	check := func(s string) bool {
		_, err := strconv.ParseInt(s, 10, 64)
		return err != nil
	}
	return makeConstraint(check, "must be an integer", message...)
}

func makeConstraint[T any](check func(T) bool, defaultMessage string, messages ...string) Constraint[T] {
	return func(t T) *ValidationError {
		if check(t) {
			return makeError(defaultMessage, messages...)
		}
		return nil
	}
}

func makeError(defaultMessage string, messages ...string) *ValidationError {
	if len(messages) != 0 {
		defaultMessage = strings.Join(messages, " ")
	}
	return NewValidationError(defaultMessage)
}
