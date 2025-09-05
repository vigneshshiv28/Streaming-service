package streaming

import (
	"math/rand"
	"sync"
	"time"
)

type Connection interface {
	Send([]byte) error
	Close()
}

type Participant struct {
	ID       string
	Name     string
	Role     string
	Conn     Connection
	RoomId   string
	Status   string
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

func (rm *RoomManager) CreateRoom(roomID string) (*Room, bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, ok := rm.Rooms[roomID]; ok {
		return room, true
	}

	room := &Room{
		ID:           roomID,
		Participants: make(map[string]*Participant),
		CreatedAt:    time.Now(),
	}

	return room, false
}

func (rm *RoomManager) GetRoom(roomID string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

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

func (rm *RoomManager) GenerateRoomID(n int) string {
	seed := time.Now().UTC().UnixNano()
	source := rand.NewSource(seed)
	rand.New(source)

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	roomID := string(b)

	return roomID
}

func (r *Room) AddParticipant(p *Participant) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Participants[p.ID] = p
}

func (r *Room) RemoveParticipant(p *Participant) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Participants, p.ID)
}

func (r *Room) Broadcast(senderID string, message []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for id, p := range r.Participants {
		if id != senderID {
			p.Conn.Send(message)
		}
	}
}
