package websocket

import (
	"log"
	"net/http"
	. "stream-server/internal/streaming"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleWebSocket(rm *RoomManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Query().Get("roomID")
		userID := r.URL.Query().Get("userID")
		role := r.URL.Query().Get("role")

		if roomID == "" || userID == "" || role == "" {
			http.Error(w, "Missing roomID or userID", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Fail to upgrade to WS")
			http.Error(w, "Something Went Wrong", http.StatusInternalServerError)
			return
		}

		room, ok := rm.GetRoom(roomID)
		if !ok {
			http.Error(w, "Room does not exist", http.StatusBadRequest)
			return
		}

		wsConnection := NewWSConnection(conn)

		p := &Participant{
			ID:       userID,
			Conn:     wsConnection,
			Role:     role,
			RoomId:   roomID,
			Status:   "active",
			SendChan: make(chan Message, 256),
			JoinedAt: time.Now(),
		}

		room.AddParticipant(p)
		defer func() {
			room.RemoveParticipant(p)
			wsConnection.Close()
		}()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			p.WritePump()
		}()

		go func() {
			defer wg.Done()
			p.ReadPump(room)
		}()

		wg.Wait()

	}
}
