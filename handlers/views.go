package handlers

import (
	"fmt"
	"html/template"
)

const ViewDir = "./web/templates"

const (
	TemplateView = "template"
	RoomView     = "room"
	HomeView     = "home"
	RoomInfoView = "room-info"
	RoomListView = "room-list"
	MessageView  = "message"
)

func FindView(name string) (*template.Template, error) {
	return template.New(name).ParseFiles(GetViewPath(name))
}

func GetViewPath(name string) string {
	return fmt.Sprintf("%s/%s.html", ViewDir, name)
}
