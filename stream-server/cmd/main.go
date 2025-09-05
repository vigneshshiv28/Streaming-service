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

	http.HandleFunc("/create-room", api.CreateRoomHandler(rm))
	http.HandleFunc("/join-room", api.JoinRoomHandler(rm))
	http.HandleFunc("/ws", ws.HandleWebSocket(rm))
	log.Println("Listening at port 8000")
	http.ListenAndServe(":8000", nil)
}
