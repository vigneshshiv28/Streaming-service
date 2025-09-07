package main

import (
	"log"
	"net/http"
	"stream-server/internal/streaming"
	api "stream-server/internal/transport/api"
	ws "stream-server/internal/transport/websocket"
)

func main() {
	rm := &streaming.RoomManager{Rooms: make(map[string]*streaming.Room)}

	withCORS := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*") // allow all (dev only!)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			h.ServeHTTP(w, r)
		}
	}

	http.HandleFunc("/create-room", withCORS(api.CreateRoomHandler(rm)))
	http.HandleFunc("/join-room", withCORS(api.JoinRoomHandler(rm)))
	http.HandleFunc("/ws", withCORS(ws.HandleWebSocket(rm)))
	log.Println("Listening at port 8000")
	http.ListenAndServe(":8000", nil)
}
