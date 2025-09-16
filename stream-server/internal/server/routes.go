package server

import (
	"stream-server/internal/transport/api"
	ws "stream-server/internal/transport/websocket"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	//Health Check

	//Room
	r.Route("/rooms", func(r chi.Router) {
		r.Post("/", api.CreateRoomHandler(s.roomManager))            // POST /rooms
		r.Post("/{roomId}/join", api.JoinRoomHandler(s.roomManager)) // POST /rooms/{id}/join
		r.Get("/{roomId}/ws", ws.HandleWebSocket(s.roomManager))
	})

	s.httpServer.Handler = r
}
