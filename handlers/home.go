package handlers

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
	"slices"

	"github.com/branow/peer-chat/config"
)

const (
	TemplateView = "template"
	RoomView     = "room"
	HomeView     = "home"
	RoomInfoView = "room-info"
	RoomListView = "room-list"
	MessageView  = "message"
	ErrorView    = "error"

	ViewDir        = "./web/templates"
	StaticFilesDir = "./web/static"
)

var (
	errNotFound       = errors.New("404")
	errInternalServer = errors.New("500")
)

var vr = NewViewResolver(ViewDir)

// HandleServeMux sets up routing for the application.
func HandleServeMux(mux *http.ServeMux) {
	// Static file handling
	fs := http.FileServer(http.Dir(StaticFilesDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Page handlers
	GetHomePage().ServeMux(mux)
	GetIcon().ServeMux(mux)
	NewRoomHandlers().HandleServeMux(mux)
}

// templateModel encapsulates data passed to the template view.
type templateModel struct {
	Content template.HTML
	Secured bool
}

func GetIcon() HandlerAdapter {
	hander := NewHandlerAdapter("/favicon.ico")
	hander.AddHandler(func(w http.ResponseWriter, r *http.Request) error {
		http.Redirect(w, r, "/static/img/favicon.ico", http.StatusMovedPermanently)
		return nil
	})
	return *hander
}

func GetHomePage() HandlerAdapter {
	handler := NewHandlerAdapter("/", "/home")

	handler.AddHandler(func(w http.ResponseWriter, r *http.Request) error {
		if !slices.Contains(handler.Paths(), r.URL.Path) {
			return errNotFound
		}

		homeHtml := bytes.NewBufferString("")
		if err := vr.ExecuteView(HomeView, homeHtml, struct{}{}); err != nil {
			return err
		}

		model := templateModel{
			Content: template.HTML(homeHtml.String()),
			Secured: config.GetConfig().Secured(),
		}

		return vr.ExecuteView(TemplateView, w, model)
	})

	handler.AddErrorHandler(
		func(err error) bool { return err == errNotFound },
		handleErrorPage(newError404),
	)
	handler.AddErrorHandler(
		func(err error) bool { return true },
		handleErrorPage(newError500),
	)
	return *handler
}
