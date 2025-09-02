package server

import (
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type SignallingMessage struct {
	Type      string `json:"type"`
	SDP       string `json:"sdp,omitempty"`
	Candidate string `json:"candidate,omitempty"`
	RoomID    string `json:"roomId,omitempty"`
}

type Participant struct {
	Host bool
	Conn *websocket.Conn
}

type Room struct {
	Mutex sync.RWMutex
	Map   map[string][]Participant
}

func (r *Room) InitRoom() {
	r.Map = make(map[string][]Participant)
}

func (r *Room) GetParticipants(roomID string) []Participant {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()
	return r.Map[roomID]
}

func (r *Room) CreateRoom() string {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	seed := time.Now().UTC().UnixNano()
	source := rand.NewSource(seed)
	rand.New(source)

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	roomID := string(b)

	r.Map[roomID] = []Participant{}
	return roomID
}

func (r *Room) InsertIntoRoom(roomID string, host bool, conn *websocket.Conn) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	p := Participant{host, conn}

	r.Map[roomID] = append(r.Map[roomID], p)

}

func (r *Room) RemoveFromRoom(roomID string, conn *websocket.Conn) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	participants := r.Map[roomID]
	for i, p := range participants {
		if p.Conn == conn {
			r.Map[roomID] = append(participants[:i], participants[i+1:]...)
			break
		}
	}

	if len(r.Map[roomID]) == 0 {
		delete(r.Map, roomID)
	}
}

func (r *Room) DeleteRoom(roomID string) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	delete(r.Map, roomID)

}
