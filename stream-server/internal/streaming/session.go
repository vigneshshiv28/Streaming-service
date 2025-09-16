package streaming

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/rs/zerolog"
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
	Role    string      `json:"role,omitempty"`
	Name    string      `json:"name,omitempty"`
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
	Rooms  map[string]*Room
	mu     sync.RWMutex
	logger *zerolog.Logger
}

func NewRoomManager(logger *zerolog.Logger) *RoomManager {
	return &RoomManager{
		Rooms:  make(map[string]*Room),
		logger: logger,
	}
}

func (rm *RoomManager) GetLogger() *zerolog.Logger {
	return rm.logger
}

func (rm *RoomManager) CreateRoom(roomID string) (*Room, bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, ok := rm.Rooms[roomID]; ok {
		rm.logger.Debug().Str("room_id", roomID).Msg("room already exists")
		return room, true
	}

	room := &Room{
		ID:           roomID,
		Participants: make(map[string]*Participant),
		CreatedAt:    time.Now(),
	}
	rm.Rooms[roomID] = room

	rm.logger.Info().Str("room_id", roomID).Msg("room created")
	return room, false
}

func (rm *RoomManager) GetRoom(roomID string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, ok := rm.Rooms[roomID]
	if !ok {
		rm.logger.Warn().Str("room_id", roomID).Msg("room not found")
		return nil, false
	}
	return room, true
}

func (rm *RoomManager) DeleteRoom(roomID string) {
	rm.mu.Lock()
	room, ok := rm.Rooms[roomID]
	if !ok {
		rm.mu.Unlock()
		rm.logger.Warn().Str("room_id", roomID).Msg("attempted to delete non-existing room")
		return
	}

	participants := make([]*Participant, 0, len(room.Participants))
	room.mu.RLock()
	for _, p := range room.Participants {
		participants = append(participants, p)
	}
	room.mu.RUnlock()
	delete(rm.Rooms, roomID)
	rm.mu.Unlock()

	for _, p := range participants {
		room.RemoveParticipant(p, rm.logger)
	}

	rm.logger.Info().Str("room_id", roomID).Msg("room deleted")
}

func (rm *RoomManager) GenerateRoomID(n int) string {
	seed := time.Now().UTC().UnixNano()
	r := rand.New(rand.NewSource(seed))

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}

	roomID := string(b)
	rm.logger.Debug().Str("room_id", roomID).Msg("generated room ID")
	return roomID
}

func (rm *RoomManager) CloseAllRooms() {
	rm.mu.Lock()
	rooms := make([]*Room, 0, len(rm.Rooms))
	for _, room := range rm.Rooms {
		rooms = append(rooms, room)
	}
	rm.Rooms = make(map[string]*Room)
	rm.mu.Unlock()

	for _, room := range rooms {
		participants := make([]*Participant, 0, len(room.Participants))
		room.mu.RLock()
		for _, p := range room.Participants {
			participants = append(participants, p)
		}
		room.mu.RUnlock()
		for _, p := range participants {
			room.RemoveParticipant(p, rm.logger)
		}
	}

	rm.logger.Info().Msg("all rooms closed")
}

func (r *Room) AddParticipant(p *Participant, logger *zerolog.Logger) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exist := r.Participants[p.ID]; exist {
		logger.Warn().Str("room_id", r.ID).Str("participant_id", p.ID).Msg("participant already exists in room")
		return fmt.Errorf("participant %s already exists in room %s", p.ID, r.ID)
	}

	r.Participants[p.ID] = p
	participantCount := len(r.Participants)

	logger.Info().Str("room_id", r.ID).Str("participant_id", p.ID).Int("participant_count", participantCount).Msg("participant added to room")
	/*
		if participantCount > 1 {
			joinMsg := Message{
				Type:    "participant_joined",
				From:    p.ID,
				Action:  "join",
				Content: fmt.Sprintf(`{"participant_count":%d,"participant_id":"%s","participant_name":"%s"}`, participantCount, p.ID, p.Name),
			}
			r.Broadcast(p.ID, joinMsg, logger)

		}*/
	logger.Info().Str("room_id", r.ID).Str("participant_id", p.ID).Int("participant_count", participantCount).Msg("notified existing participants of new join")

	return nil
}

func (r *Room) RemoveParticipant(p *Participant, logger *zerolog.Logger) {
	r.mu.Lock()
	_, ok := r.Participants[p.ID]
	if !ok {
		r.mu.Unlock()
		return
	}
	delete(r.Participants, p.ID)

	participantCount := len(r.Participants)
	isEmpty := participantCount == 0
	r.mu.Unlock()

	p.Conn.Close()
	close(p.SendChan)
	if !isEmpty {
		leaveMsg := Message{
			Type:    "participant_left",
			From:    p.ID,
			Action:  "leave",
			Content: fmt.Sprintf(`{"participant_count":%d,"participant_id":"%s","participant_name":"%s"}`, participantCount, p.ID, p.Name),
		}
		r.Broadcast(p.ID, leaveMsg, logger)
		logger.Info().Str("room_id", r.ID).Str("participant_id", p.ID).Int("participant_count", participantCount).Msg("notified remaining participants of departure")
	}

	logger.Info().Str("room_id", r.ID).Str("participant_id", p.ID).Int("participant_count", participantCount).Bool("room_empty", isEmpty).Msg("participant removed from room")
}

func (r *Room) Broadcast(senderID string, message Message, logger *zerolog.Logger) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for id, p := range r.Participants {
		if id == senderID {
			continue
		}
		select {
		case p.SendChan <- message:
		default:
			logger.Warn().Str("room_id", r.ID).Str("sender_id", senderID).Str("receiver_id", id).Msg("dropping message, send channel full")
		}
	}
}

func (r *Room) SendTo(senderID string, receiverID string, message Message, logger *zerolog.Logger) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.Participants[receiverID]
	if !ok {
		logger.Warn().Str("room_id", r.ID).Str("sender_id", senderID).Str("receiver_id", receiverID).Msg("receiver not found in room")
		return fmt.Errorf("participant %s does not exist", receiverID)
	}

	select {
	case p.SendChan <- message:
		logger.Debug().Str("room_id", r.ID).Str("sender_id", senderID).Str("receiver_id", receiverID).Str("message_type", message.Type).Msg("message sent to participant")
		return nil
	default:
		logger.Warn().Str("room_id", r.ID).Str("sender_id", senderID).Str("receiver_id", receiverID).Msg("failed to send message, channel full")
		return fmt.Errorf("not able to send the message to %s", receiverID)
	}
}

func (r *Room) GetParticipantCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Participants)
}

func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Participants) == 0
}

func (r *Room) GetParticipantList() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	participants := make([]map[string]interface{}, 0, len(r.Participants))
	for _, p := range r.Participants {
		participants = append(participants, map[string]interface{}{
			"id":        p.ID,
			"name":      p.Name,
			"role":      p.Role,
			"status":    p.Status,
			"joined_at": p.JoinedAt.Unix(),
		})
	}

	result := map[string]interface{}{
		"participant_count": len(r.Participants),
		"participants":      participants,
	}

	data, _ := json.Marshal(result)
	return string(data)
}

func (p *Participant) ReadPump(r *Room, rm *RoomManager, logger *zerolog.Logger) {
	defer func() {
		r.RemoveParticipant(p, logger)

		logger.Info().Str("room_id", r.ID).Str("participant_id", p.ID).Msg("connection closed, participant removed")
	}()

	for {
		msgBytes, err := p.Conn.Read()
		if err != nil {
			logger.Warn().Str("room_id", r.ID).Str("participant_id", p.ID).Err(err).Msg("connection read error")
			return
		}

		logger.Debug().Str("room_id", r.ID).Str("participant_id", p.ID).RawJSON("raw_message", msgBytes).Msg("received message")

		var msg Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			logger.Warn().Str("room_id", r.ID).Str("participant_id", p.ID).Err(err).Msg("invalid JSON")
			p.Conn.Send([]byte(`{"type":"error","message":"invalid JSON"}`))
			continue
		}

		if msg.Type == "" {
			logger.Warn().Str("room_id", r.ID).Str("participant_id", p.ID).Msg("missing message type")
			p.Conn.Send([]byte(`{"type":"error","message":"Missing message type"}`))
			continue
		}

		logger.Debug().Str("room_id", r.ID).Str("participant_id", p.ID).Str("message_type", msg.Type).Msg("processing message")

		switch msg.Type {
		case "chat":
			r.Broadcast(p.ID, msg, logger)
			logger.Debug().Str("room_id", r.ID).Str("participant_id", p.ID).Msg("chat message broadcasted")
		case "sdp":
			receiverID := msg.To
			if receiverID == "" {
				logger.Warn().Str("room_id", r.ID).Str("participant_id", p.ID).Msg("missing receiver ID in SDP")
				continue
			}
			if err := r.SendTo(p.ID, receiverID, msg, logger); err != nil {
				logger.Warn().Str("room_id", r.ID).Str("sender_id", p.ID).Str("receiver_id", receiverID).Err(err).Msg("failed to forward SDP")
			}
		case "ice":
			receiverID := msg.To
			if receiverID == "" {
				logger.Warn().Str("room_id", r.ID).Str("participant_id", p.ID).Msg("missing receiver ID in ICE")
				continue
			}
			if err := r.SendTo(p.ID, receiverID, msg, logger); err != nil {
				logger.Warn().Str("room_id", r.ID).Str("sender_id", p.ID).Str("receiver_id", receiverID).Err(err).Msg("failed to forward ICE")
			}
		case "get_participants":
			participantList := r.GetParticipantList()
			responseMsg := Message{
				Type:    "participant_list",
				Content: participantList,
			}
			select {
			case p.SendChan <- responseMsg:
				logger.Debug().Str("room_id", r.ID).Str("participant_id", p.ID).Msg("participant list sent")
			default:
				logger.Warn().Str("room_id", r.ID).Str("participant_id", p.ID).Msg("failed to send participant list, channel full")
			}
		case "join":
			r.Broadcast(p.ID, msg, logger)
			logger.Debug().Str("room_id", r.ID).Str("participant_id", p.ID).Msg("joining message broadcasted")

		default:
			logger.Warn().Str("room_id", r.ID).Str("participant_id", p.ID).Str("type", msg.Type).Msg("unknown message type")
			p.Conn.Send([]byte(`{"type":"error","message":"Unknown message type"}`))
		}
	}
}

func (p *Participant) WritePump(logger *zerolog.Logger) {
	for msg := range p.SendChan {
		data, err := json.Marshal(msg)
		if err != nil {
			logger.Error().Str("room_id", p.RoomId).Str("participant_id", p.ID).Err(err).Msg("failed to marshal outgoing message")
			continue
		}

		if err := p.Conn.Send(data); err != nil {
			logger.Warn().Str("room_id", p.RoomId).Str("participant_id", p.ID).Err(err).Msg("failed to send message")
			return
		}

		logger.Debug().Str("room_id", p.RoomId).Str("participant_id", p.ID).Str("message_type", msg.Type).Msg("message sent to participant")
	}
}
