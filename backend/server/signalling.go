package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var RoomManager Room

func CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	roomID := RoomManager.CreateRoom()

	type Response struct {
		RoomID string `json:"roomId"`
	}

	fmt.Println("Room created with ID:", roomID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonResponse := Response{RoomID: roomID}
	json.NewEncoder(w).Encode(jsonResponse)

}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type broadcastMessage struct {
	Message map[string]interface{}
	RoomID  string
	Client  *websocket.Conn
}

var broadcast = make(chan broadcastMessage)

func broadcaster() {
	for {
		msg := <-broadcast

		for _, client := range RoomManager.Map[msg.RoomID] {
			if client.Conn != msg.Client {
				err := client.Conn.WriteJSON(msg.Message)

				if err != nil {
					log.Fatal(err)
					client.Conn.Close()
				}
			}
		}
	}
}

func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	roomId, ok := r.URL.Query()["roomID"]

	if !ok {
		log.Println("RoomId missing")
	}

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Fatal("Fail to upgrade to websocket", err)
	}

	RoomManager.InsertIntoRoom(roomId[0], false, ws)

	go broadcaster()

	for {
		var msg broadcastMessage

		err := ws.ReadJSON(&msg.Message)

		if err != nil {
			log.Fatal("Fail to read message", err)
		}

		msg.Client = ws
		msg.RoomID = roomId[0]

		log.Println("Message received: ", msg.Message)
		broadcast <- msg
	}

}
