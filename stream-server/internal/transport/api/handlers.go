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

		var roomID string
		for {
			roomID = rm.GenerateRoomID(8)

			if _, exists := rm.Rooms[roomID]; !exists {
				break
			}
		}

		rm.CreateRoom(roomID)

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}

		guestURL := scheme + "://" + r.Host + "/join/" + roomID + "?role=guest"
		audienceURL := scheme + "://" + r.Host + "/join/" + roomID + "?role=audience"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CreateRoomResponse{
			UserID:      req.UserId,
			Name:        req.Name,
			Role:        req.Role,
			RoomID:      roomID,
			GuestURL:    guestURL,
			AudienceURL: audienceURL,
			CreatedAt:   rm.Rooms[roomID].CreatedAt.Format(`2025-09-05 18:54:00`),
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

		userID := req.UserID
		roomID := req.RoomID
		role := req.Role

		room, ok := rm.GetRoom(roomID)
		if !ok {
			http.Error(w, "Room does not exist", http.StatusBadRequest)
			return
		}

		scheme := "ws"
		if r.TLS != nil {
			scheme = "wss"
		}

		wsURL := scheme + "://" + r.Host + "/ws?room_id=" + room.ID + "&user_id=" + userID + "&role=" + role

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
