package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/branow/peer-chat/config"
	"github.com/branow/peer-chat/model"
	"github.com/branow/peer-chat/validation"
	"github.com/gorilla/websocket"
)

// WebSocket upgrader to handle WebSocket connections.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 8,
	WriteBufferSize: 1024 * 8,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// RoomHandlers manages handlers related to chat rooms.
type RoomHandlers struct {
	manager *model.RoomManager
}

func NewRoomHandlers() *RoomHandlers {
	return &RoomHandlers{
		manager: model.NewRoomManager(),
	}
}

// HandleServeMux registers all the routes handled by RoomHandlers.
func (h RoomHandlers) HandleServeMux(mux *http.ServeMux) {
	h.WsRoom().ServeMux(mux)
	h.GetRoomPage().ServeMux(mux)
	h.GetRoomList().ServeMux(mux)
	h.PostCreateRoom().ServeMux(mux)
	h.PutConnect().ServeMux(mux)
}

func (h RoomHandlers) WsRoom() HandlerAdapter {
	handler := NewHandlerAdapter("GET /ws/room/{roomId}")

	handler.AddHandler(func(w http.ResponseWriter, r *http.Request) error {
		roomIdStr := r.PathValue("roomId")
		roomId, err := strconv.ParseInt(roomIdStr, 10, 64)
		if err != nil {
			return errInternalServer
		}

		// Check if the room exits.
		_, err = h.manager.GetRoom(int(roomId))
		if err != nil {
			return err
		}

		// Upgrade HTTP connection to WebSocket.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return err
		}

		// Add client to the room and wait for interaction.
		client := model.NewClient(conn)
		_ = h.manager.AddClient(int(roomId), client)
		client.Wait()

		return nil
	})

	handler.AddErrorHandler(
		func(err error) bool { return true },
		handleErrorMessage(newError500),
	)
	return *handler
}

func (h RoomHandlers) GetRoomPage() HandlerAdapter {
	handler := NewHandlerAdapter("GET /room/{roomId}")

	handler.AddHandler(func(w http.ResponseWriter, r *http.Request) error {
		roomIdStr := r.PathValue("roomId")
		roomId, err := strconv.ParseInt(roomIdStr, 10, 64)
		if err != nil {
			return errNotFound
		}

		roomInfo, err := h.manager.GetRoom(int(roomId))
		if err != nil {
			return err
		}

		buf := bytes.NewBufferString("")
		if err := vr.ExecuteView(RoomView, buf, roomInfo); err != nil {
			return err
		}

		roomHtml := buf.String()
		model := templateModel{Content: template.HTML(roomHtml), Secured: config.GetConfig().Secured()}
		return vr.ExecuteView(TemplateView, w, model)
	})

	handler.AddErrorHandler(
		func(err error) bool { return err == errNotFound },
		handleErrorPage(newError404),
	)
	handler.AddErrorHandler(
		func(err error) bool { return errors.Is(err, model.ErrRoomDoesNotExist) },
		handleErrorPage(newError404),
	)
	handler.AddErrorHandler(
		func(err error) bool { return true },
		handleErrorPage(newError500),
	)
	return *handler
}

func (h RoomHandlers) GetRoomList() HandlerAdapter {
	handler := NewHandlerAdapter("GET /x/rooms")

	handler.AddHandler(func(w http.ResponseWriter, r *http.Request) error {
		// Fetch the template for individual room info.
		roomTmpl, err := vr.FindView(RoomInfoView)
		if err != nil {
			return err
		}

		// Fetch public rooms and render each to HTML.
		rooms := h.manager.GetPublicRooms()
		roomsHtml := []template.HTML{}
		for _, room := range rooms {
			buf := bytes.NewBufferString("")
			dto := newRoomInfoDTO(room)
			if err := roomTmpl.ExecuteTemplate(buf, RoomInfoView, dto); err != nil {
				return err
			}
			roomsHtml = append(roomsHtml, template.HTML(buf.String()))
		}

		// Render the room list view.
		model := struct{ Rooms []template.HTML }{Rooms: roomsHtml}
		return vr.ExecuteView(RoomListView, w, model)
	})

	handler.AddErrorHandler(
		func(err error) bool { return true },
		handleError(newError500),
	)
	return *handler
}

func (h RoomHandlers) PostCreateRoom() HandlerAdapter {
	handler := NewHandlerAdapter("POST /x/rooms/create")

	handler.AddHandler(func(w http.ResponseWriter, r *http.Request) error {
		name := r.PostFormValue("name")
		accessStr := r.PostFormValue("access")

		if err := validation.Validate(accessStr, "room access", validation.AnInteger()); err != nil {
			return err
		}
		access, _ := strconv.ParseInt(accessStr, 10, 64)

		room := model.NewRoomDTO(name, int(access))
		if err := room.Validate(); err != nil {
			return err
		}

		roomId, err := h.manager.CreateRoom(*room)
		if err != nil {
			return err
		}

		message := message{
			Success:     GetLocale(r).GetOr("room-was-created", "Room was created successfully"),
			RedirectURL: fmt.Sprintf("/room/%d", roomId),
		}
		return vr.ExecuteView(MessageView, w, message)
	})

	handler.AddErrorHandler(
		func(err error) bool {
			var validErr *validation.ValidationError
			return errors.As(err, &validErr) || errors.Is(err, model.ErrRoomAlreadyExists)
		},
		handleErrorMessage(newError400),
	)
	handler.AddErrorHandler(
		func(err error) bool { return true },
		handleErrorMessage(newError500),
	)

	return *handler
}

func (h RoomHandlers) PutConnect() HandlerAdapter {
	handler := NewHandlerAdapter("PUT /x/rooms/connect")

	handler.AddHandler(func(w http.ResponseWriter, r *http.Request) error {
		roomIdStr := r.PostFormValue("id")

		if err := validation.Validate(roomIdStr, "room id", validation.AnInteger()); err != nil {
			return err
		}
		roomId, _ := strconv.ParseInt(roomIdStr, 10, 64)

		if _, err := h.manager.GetRoom(int(roomId)); err != nil {
			return err
		}

		message := message{
			Success:     GetLocale(r).GetOr("room-was-found", "Room was found successfully"),
			RedirectURL: fmt.Sprintf("/room/%d", roomId),
		}
		return vr.ExecuteView(MessageView, w, message)
	})

	handler.AddErrorHandler(
		func(err error) bool {
			var validErr *validation.ValidationError
			return errors.As(err, &validErr) || errors.Is(err, model.ErrRoomDoesNotExist)
		},
		handleErrorMessage(newError400),
	)
	handler.AddErrorHandler(
		func(err error) bool { return true },
		handleErrorMessage(newError500),
	)

	return *handler
}

type roomInfoDTO struct {
	Id           int
	Name         string
	Clients      int
	CreationTime string
}

func newRoomInfoDTO(room model.RoomInfo) *roomInfoDTO {
	return &roomInfoDTO{
		Id:           room.Id,
		Name:         room.Name,
		Clients:      room.Clients,
		CreationTime: room.CreationTime.Format("15:04"),
	}
}

type message struct {
	Success     string
	Error       string
	RedirectURL string
}
