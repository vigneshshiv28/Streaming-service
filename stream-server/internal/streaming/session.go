package streaming

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"stream-server/internal/core"
	"stream-server/internal/rtc"

	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Participant struct {
	ID        string
	Name      string
	Role      string
	Conn      core.Connection
	rtcConn   core.RTCConnection
	Room      *Room
	Status    string
	SendChan  chan core.Message
	JoinedAt  time.Time
	closeOnce sync.Once
}

type TrackMeta struct {
	TrackLocal    *webrtc.TrackLocalStaticRTP
	ParticipantID string
	Kind          string
}

type Room struct {
	Name         string
	ID           string
	Participants map[string]*Participant
	trackLocals  map[string]*webrtc.TrackLocalStaticRTP
	trackMeta    map[string]TrackMeta
	CreatedAt    time.Time
	CreatedBy    string
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

func (rm *RoomManager) CreateRoom(roomID string, roomName string, createdBy string) (*Room, bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, ok := rm.Rooms[roomID]; ok {
		rm.logger.Debug().Str("room_id", roomID).Msg("room already exists")
		return room, true
	}

	room := &Room{
		Name:         roomName,
		ID:           roomID,
		Participants: make(map[string]*Participant),
		trackLocals:  make(map[string]*webrtc.TrackLocalStaticRTP),
		trackMeta:    make(map[string]TrackMeta),
		CreatedAt:    time.Now(),
		CreatedBy:    createdBy,
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
	logger.Debug().Str("room_id", r.ID).Str("participant_id", p.ID).Msg("adding participant to room")
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
	p.closeOnce.Do(func() {
		r.mu.Lock()

		if _, ok := r.Participants[p.ID]; !ok {
			r.mu.Unlock()
			return
		}

		delete(r.Participants, p.ID)
		participantCount := len(r.Participants)

		r.mu.Unlock()

		if p.rtcConn != nil {
			_ = p.rtcConn.Close(logger)
		}
		p.Conn.Close()
		close(p.SendChan)

		if participantCount > 0 {
			leaveMsg := core.Message{
				Type:    "participant_left",
				From:    p.ID,
				Action:  "leave",
				Content: fmt.Sprintf(`{"participant_count":%d,"participant_id":"%s","participant_name":"%s"}`, participantCount, p.ID, p.Name),
			}
			r.Broadcast(p.ID, leaveMsg, logger)
		}

		logger.Info().
			Str("room_id", r.ID).
			Str("participant_id", p.ID).
			Int("participant_count", participantCount).
			Msg("participant removed")
	})
}

func (r *Room) ListRTCParticipant() []*Participant {
	r.mu.Lock()
	defer r.mu.Unlock()

	participants := make([]*Participant, 0)
	for _, p := range r.Participants {
		if p.Role != "audience" && p.rtcConn != nil {
			participants = append(participants, p)
		}
	}

	return participants
}

func (r *Room) Broadcast(senderID string, message core.Message, logger *zerolog.Logger) {
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

func (r *Room) sendBackLocked(senderID string, message core.Message, logger *zerolog.Logger) error {
	p, ok := r.Participants[senderID]
	if !ok {
		return fmt.Errorf("participant not found")
	}

	select {
	case p.SendChan <- message:
		logger.Debug().Str("room_id", r.ID).Str("sender_id", senderID).Str("message_type", message.Type).Msg("message sent to participant")
		return nil
	default:
		logger.Warn().Str("room_id", r.ID).Str("sender_id", senderID).Msg("failed to send message, channel full")
		return fmt.Errorf("channel full for participant %s", senderID)
	}
}

func (r *Room) SendBack(senderID string, message core.Message, logger *zerolog.Logger) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sendBackLocked(senderID, message, logger)
}

func (r *Room) SendTo(senderID string, receiverID string, message core.Message, logger *zerolog.Logger) error {
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

		var msg core.Message
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

			if p.Role == "audience" {
				logger.Warn().Str("room_id", r.ID).Str("participant_id", p.ID).Msg("audience member sent an SDP message, ignoring")
				continue
			}
			sdp := *msg.SDP

			if sdp.Type == webrtc.SDPTypeOffer {
				var err error
				tracksMetaData := msg.IncomingTracks

				logger.Debug().Msgf("Processing %d incoming tracks", len(tracksMetaData))
				for _, trackMetaData := range tracksMetaData {

					logger.Debug().
						Str("room_id", r.ID).
						Str("participant_id", trackMetaData.ParticipantID).
						Str("track_id", trackMetaData.ClientTrackID).
						Str("track_kind", trackMetaData.Kind).
						Msg("Incoming track metadata")

					if _, exists := r.Participants[trackMetaData.ParticipantID]; !exists {
						log.Warn().Str("room_id", r.ID).Str("participant_id", trackMetaData.ParticipantID).Msg("SDP offer send by participant that doesn't exist in room")
						continue
					}

				}

				if p.rtcConn != nil {
					logger.Warn().
						Str("room_id", r.ID).
						Str("participant_id", p.ID).
						Msg("Existing RTC connection found cleaning up before creating new one")

					p.rtcConn.Close(logger)
					p.rtcConn = nil

					time.Sleep(200 * time.Millisecond)
				}

				p.rtcConn, err = rtc.NewPionRTCConnection(p, tracksMetaData, logger)

				if err != nil {
					logger.Error().Err(err).Msg("unable to create peer connection")

					errMsg := core.Message{
						Type:    "error",
						To:      p.ID,
						Content: fmt.Sprintf("Failed to handle SDP offer: %v", err),
					}

					p.Room.SendBack(p.ID, errMsg, logger)
					continue

				}

				answer, err := p.rtcConn.HandleSDPOffer(sdp, logger)
				if err != nil {
					logger.Error().Str("room_id", r.ID).Str("participant_id", p.ID).Err(err).Msg("unable to handle sdp offer")

					errMsg := core.Message{
						Type:    "error",
						To:      p.ID,
						Content: fmt.Sprintf("Failed to handle SDP offer: %v", err),
					}

					p.Room.SendBack(p.ID, errMsg, logger)
					continue

				}
				responseTrackMetaData := r.GetTracks(logger)
				for i, track := range responseTrackMetaData {
					logger.Debug().
						Int("index", i).
						Str("track_id", track.TrackID).
						Str("participant_id", track.ParticipantID).
						Str("kind", track.Kind).
						Msg("response track metadata")
				}
				responseMsg := core.Message{
					Type:           "sdp",
					SDP:            &answer,
					OutgoingTracks: responseTrackMetaData,
				}

				p.Room.SendBack(p.ID, responseMsg, logger)
				logger.Debug().Str("room_id", r.ID).Str("participant_id", p.ID).Msg("sdp answer send to the user")
				p.Room.signalPeerConnectionsLocked(logger)
			} else if sdp.Type == webrtc.SDPTypeAnswer {
				if err := p.rtcConn.HandleSDPAnswer(sdp); err != nil {
					logger.Error().Str("room_id", r.ID).Str("participant_id", p.ID).Err(err).Msg("unable to handle sdp answer")
					errMsg := core.Message{
						Type:    "error",
						To:      p.ID,
						Content: fmt.Sprintf("Failed to handle SDP answer: %v", err),
					}
					p.Room.SendBack(p.ID, errMsg, logger)
				}
			}

		case "ice":

			if p.Role == "audience" {
				continue
			}

			if p.rtcConn == nil {
				logger.Warn().Str("participant_id", p.ID).Msg("received ICE candidate before RTC connection was established, ignoring")
				continue
			}
			ice := *msg.ICE
			err := p.rtcConn.HandleICE(ice, logger)

			if err != nil {
				logger.Error().Str("room_id", r.ID).Str("participant_id", p.ID).Err(err).Msg("unable to add ICE candiate")
			}

		case "get_participants":
			participantList := r.GetParticipantList()
			responseMsg := core.Message{
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
			joiningAck := core.Message{
				Type:    "join_ack",
				To:      p.ID,
				Content: fmt.Sprintf(`{"room_id":"%s","participant_id":"%s","participant_name":"%s","participant_role":"%s"}`, r.ID, p.ID, p.Name, p.Role),
			}
			p.Room.SendBack(p.ID, joiningAck, logger)
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
			logger.Error().Str("room_id", p.Room.ID).Str("participant_id", p.ID).Err(err).Msg("failed to marshal outgoing message")
			continue
		}

		if err := p.Conn.Send(data); err != nil {
			logger.Warn().Str("room_id", p.Room.ID).Str("participant_id", p.ID).Err(err).Msg("failed to send message")
			return
		}

		logger.Debug().Str("room_id", p.Room.ID).Str("participant_id", p.ID).Str("message_type", msg.Type).Msg("message sent to participant")
	}
}

func (p *Participant) OnICECandidate(candidate *webrtc.ICECandidate, logger *zerolog.Logger) error {

	iceInit := candidate.ToJSON()
	msg := core.Message{
		Type: "ice",
		From: p.ID,
		To:   p.ID,
		Role: p.Role,
		Name: p.Name,
		ICE:  &iceInit,
	}

	p.Room.SendBack(p.ID, msg, logger)
	return nil
}
