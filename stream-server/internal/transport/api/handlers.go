package api

import (
	"encoding/json"
	"net/http"
	"stream-server/internal/streaming"
)

func CreateRoomHandler(rm *streaming.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateRoomRequest

		logger := rm.GetLogger()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn().
				Err(err).
				Str("remote_addr", r.RemoteAddr).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("failed to decode JSON request body")
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.UserId == "" || req.Name == "" {
			logger.Warn().
				Str("userId", req.UserId).
				Str("name", req.Name).
				Str("remote_addr", r.RemoteAddr).
				Msg("create room request missing required fields")
			http.Error(w, "Missing required fields: userId, name, and role", http.StatusBadRequest)
			return
		}

		var roomID string
		for {
			roomID = rm.GenerateRoomID(8)

			if _, exists := rm.CreateRoom(roomID); !exists {
				break
			}
		}

		room, exist := rm.GetRoom(roomID)

		if !exist {

			logger.Error().
				Str("roomId", roomID).
				Str("remote_addr", r.RemoteAddr).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("failed to retrieve newly created room")
			http.Error(w, "Fail to create room", http.StatusInternalServerError)
			return
		}
		logger.Info().
			Str("roomId", roomID).
			Str("remote_addr", r.RemoteAddr).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("creator_user_id", req.UserId).
			Str("creator_name", req.Name).
			Time("created_at", rm.Rooms[roomID].CreatedAt).
			Int("http_status", http.StatusOK).
			Msg("room creation request succeeded")

		httpScheme := "http"
		if r.TLS != nil {
			httpScheme = "https"
		}

		guestURL := httpScheme + "://" + r.Host + "/join/" + room.ID + "?role=guest"
		audienceURL := httpScheme + "://" + r.Host + "/join/" + room.ID + "?role=audience"
		hostURL := httpScheme + "://" + r.Host + "/join/" + room.ID + "?role=host"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CreateRoomResponse{
			UserID:      req.UserId,
			Name:        req.Name,
			RoomID:      roomID,
			Role:        "host",
			HostURL:     hostURL,
			GuestURL:    guestURL,
			AudienceURL: audienceURL,
			CreatedAt:   rm.Rooms[roomID].CreatedAt.Format(`2006-01-02 15:04:05`),
		})
	}
}

func JoinRoomHandler(rm *streaming.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := rm.GetLogger()

		var req JoinRoomRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn().
				Err(err).
				Str("remote_addr", r.RemoteAddr).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("failed to decode join room request body")
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.UserID == "" || req.RoomID == "" || req.Role == "" {
			logger.Warn().
				Str("remote_addr", r.RemoteAddr).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("missing required fields in join room request")
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		if req.Role != "host" && req.Role != "guest" && req.Role != "audience" {
			logger.Warn().
				Str("userId", req.UserID).
				Str("roomId", req.RoomID).
				Str("role", req.Role).
				Msg("invalid role in join room request")
			http.Error(w, "Invalid role", http.StatusBadRequest)
			return
		}

		userId := req.UserID
		roomId := req.RoomID
		role := req.Role

		room, ok := rm.GetRoom(roomId)
		if !ok {
			logger.Warn().
				Str("user_id", userId).
				Str("room_id", roomId).
				Msg("attempt to join non-existent room")
			http.Error(w, "Room does not exist", http.StatusBadRequest)
			return
		}

		switch role {
		case "host":
			if len(room.Participants) > 0 {
				logger.Warn().
					Str("roomId", roomId).
					Str("userId", userId).
					Msg("host already exists in room")
				http.Error(w, "Host already exists", http.StatusForbidden)
				return
			}
		case "guest":
			if len(room.Participants) >= 2 {
				logger.Warn().
					Str("roomId", roomId).
					Str("userId", userId).
					Msg("room is full for guests")
				http.Error(w, "Room is full", http.StatusForbidden)
				return
			}
		case "audience":

		}

		wsScheme := "ws"
		if r.TLS != nil {
			wsScheme = "wss"
		}

		wsURL := wsScheme + "://" + r.Host + "/rooms" + "/" + roomId +
			"/ws?userId=" + userId +
			"&role=" + role

		logger.Info().
			Str("roomId", roomId).
			Str("userId", userId).
			Str("role", role).
			Str("remote_addr", r.RemoteAddr).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("http_status", http.StatusOK).
			Msg("room join request succeeded")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(JoinRoomResponse{
			Status: "joined",
			UserID: userId,
			Role:   role,
			RoomID: roomId,
			WSURL:  wsURL,
		})
	}
}
