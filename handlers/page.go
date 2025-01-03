package handlers

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
	"slices"

	"github.com/branow/peer-chat/config"
)

func HandleServeMux(mux *http.ServeMux) {
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	GetHomePage().ServeMux(mux)
	NewRoomHandlers().HandleServeMux(mux)
}

const (
	TemplateView = "template"
	RoomView     = "room"
	HomeView     = "home"
	RoomInfoView = "room-info"
	RoomListView = "room-list"
	MessageView  = "message"
	ErrorView    = "error"
)

const DefaultViewDir = "./web/templates"

var vr = NewViewResolver(DefaultViewDir)

type templateModel struct {
	Content template.HTML
	Secured bool
}

func GetHomePage() HandlerAdapter {
	handler := NewHandlerAdapter("/", "/home")
	handler.AddHandler(func(w http.ResponseWriter, r *http.Request) error {
		url := r.URL.String()
		if !slices.Contains(handler.pathes, url) {
			return errors.New("404")
		}

		buf := bytes.NewBufferString("")
		if err := vr.ExecuteView(HomeView, buf, struct{}{}); err != nil {
			return err
		}

		homeHtml := buf.String()
		model := templateModel{Content: template.HTML(homeHtml), Secured: config.GetConfig().Secured()}
		return vr.ExecuteView(TemplateView, w, model)
	})

	handler.AddErrorHandler(func(err error) bool { return err.Error() == "404" }, handleErrorPage(newError404))
	handler.AddErrorHandler(func(err error) bool { return true }, handleErrorPage(newError500))
	return *handler
}
