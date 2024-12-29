package valid

import (
	"fmt"
	"strings"
)

func Validate[T any](obj T, field string, constraints ...Constraint[T]) error {
	for _, constraint := range constraints {
		if err := constraint(obj); err != nil {
			err.Field = field
			return err
		}
	}
	return nil
}

type Constraint[T any] func(T) *ValidationError

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

func NotEmpty(messages ...string) Constraint[string] {
	check := func(s string) bool { return len(s) == 0 }
	return makeConstraint(check, "is mandatory", messages...)
}

func NotBlank(message ...string) Constraint[string] {
	check := func(s string) bool { return len(strings.TrimSpace(s)) == 0 }
	return makeConstraint(check, "is mandatory", message...)
}

func NotLongerThan(value int, messages ...string) Constraint[string] {
	check := func(s string) bool { return len(s) > value }
	return makeConstraint(check, "is too long", messages...)
}

func NotShorterThan(value int, messages ...string) Constraint[string] {
	check := func(s string) bool { return len(s) < value }
	return makeConstraint(check, "is too short", messages...)
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
