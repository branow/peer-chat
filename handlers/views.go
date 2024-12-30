package handlers

import (
	"fmt"
	"html/template"
	"io"
)

const ViewDir = "./web/templates"

const (
	TemplateView = "template"
	RoomView     = "room"
	HomeView     = "home"
	RoomInfoView = "room-info"
	RoomListView = "room-list"
	MessageView  = "message"
	ErrorView    = "error"
)

func ExecuteView(name string, w io.Writer, model any) error {
	tmpl, err := FindView(name)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, name, model)
}

func FindView(name string) (*template.Template, error) {
	return template.New(name).ParseFiles(GetViewPath(name))
}

func GetViewPath(name string) string {
	return fmt.Sprintf("%s/%s.html", ViewDir, name)
}
