package model

import (
	"errors"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/branow/peer-chat/validation"
)

var (
	ErrRoomAlreadyExists = errors.New("room already exists")
	ErrRoomDoesNotExist  = errors.New("room does not exist")
)

// RoomManager holdes and manages peer-to-peer connections.
type RoomManager struct {
	rooms map[int]*room
	mutex sync.RWMutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: map[int]*room{},
	}
}

func (m *RoomManager) GetRoom(roomId int) (RoomInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if room, ok := m.rooms[roomId]; ok {
		return *newRoomInfo(*room), nil
	}
	return RoomInfo{}, ErrRoomDoesNotExist
}

func (m *RoomManager) GetPublicRooms() []RoomInfo {
	m.removeEmptyRooms()

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	rooms := []RoomInfo{}
	for _, room := range m.rooms {
		if room.access == public {
			rooms = append(rooms, *newRoomInfo(*room))
		}
	}

	// Sort rooms by creation date, newest first
	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i].CreationTime.After(rooms[j].CreationTime)
	})
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
	slog.Info("Created room:", "room-id", room.Id())
	return room.Id(), nil
}

func (m *RoomManager) removeEmptyRooms() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, room := range m.rooms {
		// Check wheather the room is empty and remove it if so
		if room.GetClients() == 0 {
			delete(m.rooms, room.Id())
			slog.Info("Removed room as empty:", "room-id", room.Id())
		}
	}

}

func (m *RoomManager) removeRoom(roomId int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.rooms, roomId)
	slog.Info("Removed room:", "room-id", roomId)
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

// RoomDTO represents data required to create a room.
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
	err := validation.Validate(r.name, "room name",
		validation.NotBlank(),
		validation.NotShorterThan(3),
		validation.NotLongerThan(50))
	if err != nil {
		return err
	}
	err = validation.Validate(r.access, "room access", validation.Equal([]int{public, private}))
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

// RoomInfo represents public information about a room.
type RoomInfo struct {
	Id           int
	Name         string
	Clients      int
	CreationTime time.Time
}

func newRoomInfo(room room) *RoomInfo {
	return &RoomInfo{
		Id:           room.Id(),
		Name:         room.name,
		Clients:      room.GetClients(),
		CreationTime: room.creationTime,
	}
}
