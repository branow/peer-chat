package handlers

import (
	"log/slog"
	"net/http"
	"os"
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

// GetLocalizor returns the initialized Localizor instance.
func GetLocalizor() *i18n.Localizor {
	return localizor
}

// GetLocale retrieves the best matching Locale for the request based on
// the Accept-Language header.
func GetLocale(r *http.Request) i18n.Locale {
	langs := append(getAcceptLanguages(r), DefaultLang)
	locale, err := GetLocalizor().GetLocale(langs...)
	if err != nil {
		slog.Error("Get locale:", "error", err, "langs", langs)
		os.Exit(1)
	}
	return locale
}

func getAcceptLanguages(r *http.Request) []string {
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
