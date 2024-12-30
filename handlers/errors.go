package handlers

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"
)

type errorModel struct {
	Status  int
	Title   string
	Message string
	Cause   string
	GoHome  bool
}

type newErrorModel func(error) errorModel

func newError500(err error) errorModel {
	return errorModel{
		Status:  http.StatusInternalServerError,
		Title:   "Internal Server Error",
		Message: "Oops, something went wrong. Try to refresh page or feel free to contract us if the problem persists.",
		Cause:   err.Error(),
	}
}

func newError404(err error) errorModel {
	return errorModel{
		Status:  http.StatusNotFound,
		Title:   "Page Not Found",
		Message: "The page you are looking for might have been removed, had its name changed or is temporarily unavailable.",
		GoHome:  true,
		Cause:   err.Error(),
	}
}

func newError400(err error) errorModel {
	return errorModel{
		Status:  http.StatusBadRequest,
		Title:   "Bad Request",
		Message: err.Error(),
		Cause:   err.Error(),
	}
}

func handleErrorMessage(newErrorModel newErrorModel) HandleError {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		errModel := newErrorModel(err)
		w.WriteHeader(errModel.Status)
		slog.Debug("Error Response", "sratus", errModel.Status, "url", r.URL, "error", errModel.Cause)

		messageModel := message{Error: errModel.Message}
		if err := ExecuteView(MessageView, w, messageModel); err != nil {
			logError(errModel.Status, r.URL.String(), err)
		}
	}
}

func handleErrorPage(newErrorModel newErrorModel) HandleError {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		errModel := newErrorModel(err)
		w.WriteHeader(errModel.Status)
		slog.Debug("Error Response", "sratus", errModel.Status, "url", r.URL, "error", errModel.Cause)

		buf := bytes.NewBufferString("")
		if err := ExecuteView(ErrorView, buf, errModel); err != nil {
			logError(errModel.Status, r.URL.String(), err)
		}

		errorHtml := template.HTML(buf.String())
		model := templateModel{Content: errorHtml}
		if err := ExecuteView(TemplateView, w, model); err != nil {
			logError(errModel.Status, r.URL.String(), err)
		}
	}
}

func handleError(newErrorModel newErrorModel) HandleError {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		errModel := newErrorModel(err)
		slog.Debug("Error Response", "sratus", errModel.Status, "url", r.URL, "error", errModel.Cause)
		if err := ExecuteView(ErrorView, w, errModel); err != nil {
			logError(errModel.Status, r.URL.String(), err)
		}
	}
}

func logError(status int, url string, err error) {
	slog.Error("Error Response", "status", status, "url", url, "error", err)
}
