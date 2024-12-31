package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/branow/peer-chat/i18n"
)

const DefaultLang = "en"

var localizor *i18n.Localizor

func init() {
	var err error
	if localizor, err = i18n.NewLocalizor("./locales"); err != nil {
		slog.Error("Init localizor:", "error", err)
	}
}

func GetLocalizor() *i18n.Localizor {
	return localizor
}

func GetLocale(r *http.Request) i18n.Locale {
	langs := append(getAcceptLanuages(r), DefaultLang)
	locale, err := GetLocalizor().GetLocale(langs...)
	if err != nil {
		panic(err)
	}
	return locale
}

func getAcceptLanuages(r *http.Request) []string {
	header := r.Header.Get("Accept-Language")
	locales := strings.Split(header, ",")

	langs := []string{}
outer:
	for _, locale := range locales {
		l := strings.Split(locale, ";")[0]
		newLang := strings.Split(l, "-")[0]

		for _, lang := range langs {
			if lang == newLang {
				continue outer
			}
		}
		langs = append(langs, newLang)
	}

	return langs
}
