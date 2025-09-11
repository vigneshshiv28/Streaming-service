package api

import (
	"encoding/json"
	"net/http"
	"stream-server/internal/streaming"
)

func CreateRoomHandler(rm *streaming.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateRoomRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.UserId == "" || req.Name == "" || req.Role == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
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
			http.Error(w, "Fail to create room", http.StatusInternalServerError)
			return
		}

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
			Role:        req.Role,
			RoomID:      roomID,
			HostURL:     hostURL,
			GuestURL:    guestURL,
			AudienceURL: audienceURL,
			CreatedAt:   rm.Rooms[roomID].CreatedAt.Format(`2006-01-02 15:04:05`),
		})
	}
}

func JoinRoomHandler(rm *streaming.RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req JoinRoomRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invaild request", http.StatusBadRequest)
			return
		}

		if req.UserID == "" || req.RoomID == "" || req.Role == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		if req.Role != "host" && req.Role != "guest" && req.Role != "audience" {
			http.Error(w, "Invalid Request", http.StatusBadRequest)
		}

		userID := req.UserID
		roomID := req.RoomID
		role := req.Role

		room, ok := rm.GetRoom(roomID)
		if !ok {
			http.Error(w, "Room does not exist", http.StatusBadRequest)
			return
		}

		switch role {
		case "host":
			if len(room.Participants) > 0 {
				http.Error(w, "Host already exists", http.StatusForbidden)
				return
			}
		case "guest":
			if len(room.Participants) >= 2 {
				http.Error(w, "Room is full", http.StatusForbidden)
				return
			}
		case "audience":

		default:
			http.Error(w, "Invalid Request", http.StatusBadRequest)
			return

		}

		wsScheme := "ws"
		if r.TLS != nil {
			wsScheme = "wss"
		}

		wsURL := wsScheme + "://" + r.Host + "/ws?room_id=" + room.ID + "&user_id=" + userID + "&role=" + role

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(JoinRoomResponse{
			Status: "joined",
			UserID: userID,
			Role:   role,
			RoomID: roomID,
			WSURL:  wsURL,
		})

	}
}
