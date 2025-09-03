package streaming

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Participant struct {
	ID       string
	Name     string
	Role     string
	Conn     websocket.Conn
	RoomId   string
	JoinedAt time.Time
}

type Room struct {
	ID           string
	Participants map[string]*Participant
	CreatedAt    time.Time
	mu           sync.RWMutex
}

type RoomManager struct {
	Rooms map[string]*Room
	mu    sync.RWMutex
}

func (rm *RoomManager) CreateRoom(roomID string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, ok := rm.Rooms[roomID]; ok {
		return room
	}

	room := &Room{
		ID:           roomID,
		Participants: make(map[string]*Participant),
		CreatedAt:    time.Now(),
	}

	return room
}

func (rm *RoomManager) GetRoom(roomID string) (*Room, bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, ok := rm.Rooms[roomID]; ok {
		return room, true
	} else {
		return nil, false
	}
}

func (rm *RoomManager) DeleteRoom(roomID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, ok := rm.Rooms[roomID]; ok {
		for _, p := range room.Participants {
			p.Conn.Close()
		}
		delete(rm.Rooms, roomID)
	}
}

func (r *Room) AddParticipant(p *Participant) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Participants[p.ID] = p
}

func (r *Room) DeleteRoom(p *Participant) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Participants, p.ID)
}
