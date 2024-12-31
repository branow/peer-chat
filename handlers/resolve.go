package handlers

import "strings"

type HasI18NKey interface {
	GetI18NKey() string
}

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
