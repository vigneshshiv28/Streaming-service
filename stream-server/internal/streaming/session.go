package streaming

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Connection interface {
	Send([]byte) error
	Close()
	Read() ([]byte, error)
}

type Message struct {
	Type    string      `json:"type"`
	From    string      `json:"from,omitempty"`
	To      string      `json:"to,omitempty"`
	Content string      `json:"content,omitempty"`
	SDP     interface{} `json:"sdp,omitempty"`
	ICE     interface{} `json:"ice,omitempty"`
	Action  string      `json:"action,omitempty"`
}

type Participant struct {
	ID       string
	Name     string
	Role     string
	Conn     Connection
	RoomId   string
	Status   string
	SendChan chan Message
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
	rm.Rooms[roomID] = room

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

func (p *Participant) ReadPump(r *Room) {
	defer func() {
		r.RemoveParticipant(p)
	}()

	for {
		msgBytes, error := p.Conn.Read()
		if error != nil {
			fmt.Printf("Fail to read the message")
		}

		var msg Message

		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			fmt.Printf("Invalid Message %s", msg)
		}

		switch msg.Type {
		case "chat":
			r.Broadcast(p.ID, msg)
		case "sdp":
			_, ok := msg.SDP.(map[string]interface{})
			if !ok {
				fmt.Printf("Invalid SDP Message %s", msg)
				continue
			}

			receiverID := msg.To
			if receiverID == "" {
				fmt.Printf("Missing receiver ID in the SDP message %s", msg)
				continue
			}
			if err := r.SendTo(p.ID, receiverID, msg); err != nil {
				fmt.Printf("Fail to send the message to %s", receiverID)
			}
		case "ice":
			_, ok := msg.ICE.(map[string]interface{})
			if !ok {
				fmt.Printf("Invalid ICE Message %s", msg)
				continue
			}

			receiverID := msg.To
			if receiverID == "" {
				fmt.Printf("Missing receiver ID in the ICE message %s", msg)
				continue
			}

			if err := r.SendTo(p.ID, receiverID, msg); err != nil {
				fmt.Printf("Fail to send the message to %s", receiverID)
			}
		}

	}
}

func (p *Participant) WritePump() {
	defer p.Conn.Close()

	for msg := range p.SendChan {
		data, err := json.Marshal(msg)

		if err != nil {
			fmt.Println("Marshalling Error:", err)
		}

		if err := p.Conn.Send(data); err != nil {
			fmt.Println("Error in sending message:", err)
			return
		}

	}

}

func (r *Room) Broadcast(senderID string, message Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for id, p := range r.Participants {
		if id == senderID {
			continue
		}

		select {
		case p.SendChan <- message:
		default:
			log.Printf("Not able to send the message to %s", id)
		}
	}
}

func (r *Room) SendTo(senderID string, receiverID string, message Message) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.Participants[receiverID]
	if !ok {
		return fmt.Errorf("Participant %s does not exist", receiverID)
	}

	select {
	case p.SendChan <- message:
		return nil
	default:
		return fmt.Errorf("not able to send the message %s", receiverID)
	}

}
