package websocket

import (
	"net/http"
	"stream-server/internal/core"
	. "stream-server/internal/streaming"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleWebSocket(rm *RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, "roomId")
		userID := r.URL.Query().Get("userId")
		role := r.URL.Query().Get("role")

		logger := rm.GetLogger()

		if roomID == "" || userID == "" || role == "" {
			logger.Warn().
				Str("room_id", roomID).
				Str("user_id", userID).
				Str("role", role).
				Str("remote_addr", r.RemoteAddr).
				Msg("WebSocket connection attempt with missing parameters")
			http.Error(w, "Missing roomID or userID", http.StatusBadRequest)
			return
		}

		room, ok := rm.GetRoom(roomID)

		if existing, exists := room.Participants[userID]; exists {
			logger.Warn().
				Str("room_id", roomID).
				Str("user_id", userID).
				Str("role", role).
				Str("remote_addr", r.RemoteAddr).
				Msg("User already have joined the room")
			room.RemoveParticipant(existing, logger)

		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error().
				Str("room_id", roomID).
				Str("user_id", userID).
				Str("remote_addr", r.RemoteAddr).
				Err(err).
				Msg("failed to upgrade to WebSocket")
			http.Error(w, "Something Went Wrong", http.StatusInternalServerError)
			return
		}

		logger.Debug().
			Str("room_id", roomID).
			Str("user_id", userID).
			Str("role", role).
			Str("remote_addr", r.RemoteAddr).
			Msg("WebSocket connection upgraded successfully")

		if !ok {
			logger.Warn().
				Str("room_id", roomID).
				Str("user_id", userID).
				Msg("attempted to join non-existent room")
			conn.Close()
			http.Error(w, "Room does not exist", http.StatusBadRequest)
			return
		}

		wsConnection := NewWSConnection(conn)

		p := &Participant{
			ID:       userID,
			Conn:     wsConnection,
			Role:     role,
			Room:     room,
			Status:   "active",
			SendChan: make(chan core.Message, 256),
			JoinedAt: time.Now(),
		}

		logger.Info().Str("room_id", roomID).Str("user_id", userID).Str("role", role).Msg("attempting to add participant to room")

		if err := room.AddParticipant(p, logger); err != nil {
			logger.Error().Str("room_id", roomID).Str("user_id", userID).Err(err).Msg("failed to add participant to room")
			wsConnection.Close()
			return
		}

		defer func() {
			logger.Info().Str("room_id", roomID).Str("user_id", userID).Msg("cleaning up WebSocket connection")
			//room.RemoveParticipant(p, logger)
		}()

		var wg sync.WaitGroup
		wg.Add(2)

		logger.Debug().Str("room_id", roomID).Str("user_id", userID).Msg("starting WebSocket read/write pumps")

		go func() {
			defer wg.Done()
			defer logger.Debug().Str("room_id", roomID).Str("user_id", userID).Msg("write pump terminated")
			p.WritePump(logger)
		}()

		go func() {
			defer wg.Done()
			defer logger.Debug().Str("room_id", roomID).Str("user_id", userID).Msg("read pump terminated")
			p.ReadPump(room, rm, logger)
		}()

		wg.Wait()

		logger.Info().Str("room_id", roomID).Str("user_id", userID).Msg("WebSocket connection handler completed")
	}
}
