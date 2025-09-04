package main

import (
	"log"
	"net/http"
	"stream-server/internal/streaming"
	ws "stream-server/internal/transport/websocket"
)

func main() {
	rm := &streaming.RoomManager{Rooms: make(map[string]*streaming.Room)}

	http.HandleFunc("/ws", ws.HandleWebSocket(rm))
	log.Println("Listening at port 8000")
	http.ListenAndServe(":8000", nil)
}
