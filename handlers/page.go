package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
)

func HandleServeMux(mux *http.ServeMux) {
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	GetHomePage().ServeMux(mux)
	NewRoomHandlers().HandleServeMux(mux)
}

type templateModel struct {
	Content template.HTML
}

func GetHomePage() HandlerAdapter {
	handler := NewHandlerAdapter("GET /home")
	handler.AddHandler(func(w http.ResponseWriter, r *http.Request) error {
		homeTmpl, err := FindView(HomeView)
		if err != nil {
			return err
		}

		buf := bytes.NewBufferString("")
		if err := homeTmpl.ExecuteTemplate(buf, HomeView, struct{}{}); err != nil {
			return err
		}

		homeHtml := buf.String()
		model := templateModel{Content: template.HTML(homeHtml)}
		return executeTemplateView(w, model)
	})
	handler.AddErrorHandler(func(err error) bool { return true }, handleError500)
	return *handler
}

func handleError500(err error, w http.ResponseWriter, r *http.Request) {
	slog.Debug("Response Status 500", "error", err, "url", r.URL)
	w.WriteHeader(http.StatusInternalServerError)
	htmlContent := fmt.Sprintf("Error 500: %v", err)
	model := templateModel{Content: template.HTML(htmlContent)}
	if err := executeTemplateView(w, model); err != nil {
		slog.Error("Handler Error 500", "error", err)
	}
}

func handleError404(err error, w http.ResponseWriter, r *http.Request) {
	slog.Debug("Response Status 404", "error", err, "url", r.URL)
	w.WriteHeader(http.StatusNotFound)
	htmlContent := fmt.Sprintf("Error 404: %v", err)
	model := templateModel{Content: template.HTML(htmlContent)}
	if err := executeTemplateView(w, model); err != nil {
		slog.Error("Handler Error 404", "error", err)
	}
}

func executeTemplateView(w io.Writer, model templateModel) error {
	tmpl, err := FindView(TemplateView)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, model)
}
