package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/branow/peer-chat/model"
	"github.com/branow/peer-chat/valid"
	"golang.org/x/net/websocket"
)

type RoomHandlers struct {
	manager *model.RoomManager
}

func NewRoomHandlers() *RoomHandlers {
	return &RoomHandlers{
		manager: model.NewRoomManager(),
	}
}

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
			return errors.New("400")
		}

		_, err = h.manager.GetRoom(int(roomId))
		if err != nil {
			return err
		}

		websocket.Handler(func(c *websocket.Conn) {
			client := model.NewClient(c)
			_ = h.manager.AddClient(int(roomId), client)
			client.Wait()
		}).ServeHTTP(w, r)

		return err
	})

	handler.AddErrorHandler(func(err error) bool { return true }, handleErrorMessage(newError500))
	return *handler
}

func (h RoomHandlers) GetRoomPage() HandlerAdapter {
	handler := NewHandlerAdapter("GET /room/{roomId}")

	handler.AddHandler(func(w http.ResponseWriter, r *http.Request) error {
		roomIdStr := r.PathValue("roomId")
		roomId, err := strconv.ParseInt(roomIdStr, 10, 64)
		if err != nil {
			return errors.New("404")
		}

		roomInfo, err := h.manager.GetRoom(int(roomId))
		if err != nil {
			return err
		}

		buf := bytes.NewBufferString("")
		if err := ExecuteView(RoomView, buf, roomInfo); err != nil {
			return err
		}

		roomHtml := buf.String()
		model := templateModel{Content: template.HTML(roomHtml)}
		return ExecuteView(TemplateView, w, model)
	})

	handler.AddErrorHandler(func(err error) bool { return err.Error() == "404" }, handleErrorPage(newError404))
	handler.AddErrorHandler(func(err error) bool { return errors.Is(err, model.ErrRoomDoesNotExist) }, handleErrorPage(newError404))
	handler.AddErrorHandler(func(err error) bool { return true }, handleErrorPage(newError500))
	return *handler
}

func (h RoomHandlers) GetRoomList() HandlerAdapter {
	handler := NewHandlerAdapter("GET /x/rooms")
	handler.AddHandler(func(w http.ResponseWriter, r *http.Request) error {
		roomTmpl, err := FindView(RoomInfoView)
		if err != nil {
			return err
		}

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

		model := struct{ Rooms []template.HTML }{Rooms: roomsHtml}
		return ExecuteView(RoomListView, w, model)
	})
	handler.AddErrorHandler(func(err error) bool { return true }, handleError(newError500))
	return *handler
}

func (h RoomHandlers) PostCreateRoom() HandlerAdapter {
	handler := NewHandlerAdapter("POST /x/rooms/create")

	handler.AddHandler(func(w http.ResponseWriter, r *http.Request) error {
		name := r.PostFormValue("name")
		accessStr := r.PostFormValue("access")

		access, err := strconv.ParseInt(accessStr, 10, 32)
		if err != nil {
			return valid.NewValidationError("room access is not an integer")
		}

		room := model.NewRoomDTO(name, int(access))
		if err := room.Validate(); err != nil {
			return err
		}

		roomId, err := h.manager.CreateRoom(*room)
		if err != nil {
			return err
		}

		message := message{
			Success:     "Room was created successfully",
			RedirectURL: fmt.Sprintf("/room/%d", roomId),
		}
		return ExecuteView(MessageView, w, message)
	})

	handler.AddErrorHandler(func(err error) bool {
		var validErr *valid.ValidationError
		return errors.As(err, &validErr) || errors.Is(err, model.ErrRoomAlreadyExists)
	}, handleErrorMessage(newError400))
	handler.AddErrorHandler(func(err error) bool { return true }, handleErrorMessage(newError500))

	return *handler
}

func (h RoomHandlers) PutConnect() HandlerAdapter {
	handler := NewHandlerAdapter("PUT /x/rooms/connect")

	handler.AddHandler(func(w http.ResponseWriter, r *http.Request) error {
		roomIdStr := r.PostFormValue("id")
		roomId, err := strconv.ParseInt(roomIdStr, 10, 64)
		if err != nil {
			return valid.NewValidationError("room id must be an integer")
		}

		if _, err := h.manager.GetRoom(int(roomId)); err != nil {
			return err
		}

		message := message{
			Success:     "Room was found successfully",
			RedirectURL: fmt.Sprintf("/room/%d", roomId),
		}
		return ExecuteView(MessageView, w, message)
	})

	handler.AddErrorHandler(func(err error) bool {
		var validErr *valid.ValidationError
		return errors.As(err, &validErr) || errors.Is(err, model.ErrRoomDoesNotExist)
	}, handleErrorMessage(newError400))
	handler.AddErrorHandler(func(err error) bool { return true }, handleErrorMessage(newError500))

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
