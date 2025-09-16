package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"stream-server/internal/streaming"

	"github.com/rs/zerolog"
)

type Server struct {
	httpServer  *http.Server
	logger      *zerolog.Logger
	roomManager *streaming.RoomManager
}

func NewServer(logger *zerolog.Logger, rm *streaming.RoomManager) *Server {
	return &Server{
		logger:      logger,
		roomManager: rm,
	}

}

func (s *Server) SetupServer(addr string) {
	s.httpServer = &http.Server{
		Addr: ":" + addr,
	}

}

func (s *Server) StartServer() error {
	if s.httpServer == nil {
		return errors.New("http server is not intialized")
	}

	s.logger.Info().Str("port", s.httpServer.Addr).Msg("started HTTP server")

	return s.httpServer.ListenAndServe()
}

func (s *Server) StopServer(ctx context.Context) error {
	s.roomManager.CloseAllRooms()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return nil

}
