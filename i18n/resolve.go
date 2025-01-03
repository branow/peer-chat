package i18n

import "strings"

// HasI18NKey is an interface for errors or objects that
// provide an i18 key.
type HasI18NKey interface {
	GetI18NKey() string
}

// ResolveI18NKeyOfError resolves the i18n key for the given error.
// If the error implements HasI18NKey interface, it returns
// the result of GetI18NKey(). Otherwise a key using the error's
// string representation.
func ResolveI18NKeyOfError(err error) string {
	if hasI18NKey, ok := err.(HasI18NKey); ok {
		return hasI18NKey.GetI18NKey()
	}
	return toI18NKey(err.Error())
}

func toI18NKey(str string) string {
	str = strings.TrimSpace(str)
	str = strings.ToLower(str)
	str = strings.ReplaceAll(str, " ", "-")
	return str
}
