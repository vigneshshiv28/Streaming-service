package server

import (
	"stream-server/internal/transport/api"
	ws "stream-server/internal/transport/websocket"

	"github.com/go-chi/chi/v5"
)

func (s *Server) RegisterRoutes() {
	r := chi.NewRouter()

	//Health Check

	//Room
	r.Route("/rooms", func(r chi.Router) {
		r.Post("/", api.CreateRoomHandler(s.roomManager))        // POST /rooms
		r.Post("/{id}/join", api.JoinRoomHandler(s.roomManager)) // POST /rooms/{id}/join
		r.Get("/{id}/ws", ws.HandleWebSocket(s.roomManager))
	})

	s.httpServer.Handler = r
}
