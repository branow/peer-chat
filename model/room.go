package model

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/branow/peer-chat/valid"
)

var (
	ErrRoomAlreadyExists = errors.New("room already exists")
	ErrRoomDoesNotExist  = errors.New("room does not exist")
)

type RoomManager struct {
	rooms map[int]*room
	mutex sync.RWMutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: map[int]*room{},
	}
}

func (m *RoomManager) GetRoom(roomId int) (roomInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if room, ok := m.rooms[roomId]; ok {
		return *newRoomInfo(*room), nil
	}
	return roomInfo{}, ErrRoomDoesNotExist
}

func (m *RoomManager) GetPublicRooms() []roomInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	rooms := []roomInfo{}
	for _, room := range m.rooms {
		if room.access == public {
			rooms = append(rooms, *newRoomInfo(*room))
		}
	}
	return rooms
}

func (m *RoomManager) CreateRoom(dto RoomDTO) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, room := range m.rooms {
		if room.name == dto.name {
			return 0, ErrRoomAlreadyExists
		}
	}

	room := newRoom(dto.name, dto.access)
	room.SetOnEmptyConnection(func() { m.removeRoom(room.id) })
	m.rooms[room.Id()] = room
	return room.Id(), nil
}

func (m *RoomManager) removeRoom(roomId int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.rooms, roomId)
	slog.Debug("Remove room", "room-id", roomId)
}

func (m *RoomManager) AddClient(roomId int, client *Client) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if room, ok := m.rooms[roomId]; ok {
		room.AddClient(client)
	}
	return ErrRoomDoesNotExist
}

const (
	private = iota
	public
)

type RoomDTO struct {
	name   string
	access int
}

func NewRoomDTO(name string, access int) *RoomDTO {
	return &RoomDTO{
		name:   name,
		access: access,
	}
}

func (r RoomDTO) Validate() error {
	err := valid.Validate(r.name, "room name",
		valid.NotBlank(),
		valid.NotShorterThan(3),
		valid.NotLongerThan(50))
	if err != nil {
		return err
	}
	err = valid.Validate(r.access, "room access", valid.Equal([]int{public, private}))
	if err != nil {
		return err
	}
	return nil
}

type room struct {
	*PeerConnection
	name         string
	access       int
	creationTime time.Time
}

func newRoom(name string, access int) *room {
	return &room{
		PeerConnection: NewPeerConnection(),
		name:           name,
		access:         access,
		creationTime:   time.Now(),
	}
}

type roomInfo struct {
	Id           int
	Name         string
	Clients      int
	CreationTime time.Time
}

func newRoomInfo(room room) *roomInfo {
	return &roomInfo{
		Id:           room.Id(),
		Name:         room.name,
		Clients:      room.GetClients(),
		CreationTime: room.creationTime,
	}
}
