package handlers

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/branow/peer-chat/i18n"
)

type errorModel struct {
	Status           int
	Title            string
	Message          string
	Cause            string
	GoHome           bool
	localizationKeys map[string]string
}

func (e *errorModel) localize(locale i18n.Locale) {
	fieldSetters := map[string]i18n.Setter{
		"Title":   func(s string) { e.Title = s },
		"Message": func(s string) { e.Message = s },
	}

	keySetters := map[string]i18n.Setter{}
	for fieldName, setter := range fieldSetters {
		if key, ok := e.localizationKeys[fieldName]; ok {
			keySetters[key] = setter
		}
	}

	if err := locale.LocalizeFields(keySetters); err != nil {
		slog.Error("Error model localization:", "error", err)
	}
}

type newErrorModel func(error) errorModel

func newError500(err error) errorModel {
	return errorModel{
		Status:  http.StatusInternalServerError,
		Title:   "Internal Server Error",
		Message: "Oops, something went wrong. Try to refresh page or feel free to contract us if the problem persists.",
		Cause:   err.Error(),
		localizationKeys: map[string]string{
			"Title":   "error-500-title",
			"Message": "error-500-message",
		},
	}
}

func newError404(err error) errorModel {
	return errorModel{
		Status:  http.StatusNotFound,
		Title:   "Page Not Found",
		Message: "The page you are looking for might have been removed, had its name changed or is temporarily unavailable.",
		GoHome:  true,
		Cause:   err.Error(),
		localizationKeys: map[string]string{
			"Title":   "error-404-title",
			"Message": "error-404-message",
		},
	}
}

func newError400(err error) errorModel {
	return errorModel{
		Status:  http.StatusBadRequest,
		Title:   "Bad Request",
		Message: err.Error(),
		Cause:   err.Error(),
		localizationKeys: map[string]string{
			"Title":   "error-400-title",
			"Message": ResolveI18NKeyOfError(err),
		},
	}
}

func handleErrorMessage(newErrorModel newErrorModel) HandleError {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		errModel := newErrorModel(err)
		errModel.localize(GetLocale(r))
		w.WriteHeader(errModel.Status)
		slog.Debug("Error Response", "status", errModel.Status, "url", r.URL, "error", errModel.Cause)

		messageModel := message{Error: errModel.Message}
		if err := vr.ExecuteView(MessageView, w, messageModel); err != nil {
			logError(errModel.Status, r.URL.String(), err)
		}
	}
}

func handleErrorPage(newErrorModel newErrorModel) HandleError {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		errModel := newErrorModel(err)
		errModel.localize(GetLocale(r))
		w.WriteHeader(errModel.Status)
		slog.Debug("Error Response", "status", errModel.Status, "url", r.URL, "error", errModel.Cause)

		buf := bytes.NewBufferString("")
		if err := vr.ExecuteView(ErrorView, buf, errModel); err != nil {
			logError(errModel.Status, r.URL.String(), err)
		}

		errorHtml := template.HTML(buf.String())
		model := templateModel{Content: errorHtml}
		if err := vr.ExecuteView(TemplateView, w, model); err != nil {
			logError(errModel.Status, r.URL.String(), err)
		}
	}
}

func handleError(newErrorModel newErrorModel) HandleError {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		errModel := newErrorModel(err)
		errModel.localize(GetLocale(r))
		slog.Debug("Error Response", "status", errModel.Status, "url", r.URL, "error", errModel.Cause)
		if err := vr.ExecuteView(ErrorView, w, errModel); err != nil {
			logError(errModel.Status, r.URL.String(), err)
		}
	}
}

func logError(status int, url string, err error) {
	slog.Error("Error Response", "status", status, "url", url, "error", err)
}
