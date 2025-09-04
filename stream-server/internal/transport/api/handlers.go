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
	}
}
