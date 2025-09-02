package main

import (
	"backend/server"
	"fmt"
	"log"
	"net/http"
)

func main() {

	server.RoomManager.InitRoom()

	http.HandleFunc("/create", server.CreateRoomHandler)
	http.HandleFunc("/join", server.JoinRoomHandler)

	log.Println("starting Server on  Port 8000")
	fmt.Println(" ")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal((err))
	}

}
