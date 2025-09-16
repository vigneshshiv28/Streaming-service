package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WSConnection struct {
	conn *websocket.Conn
	mu   sync.RWMutex
}

func NewWSConnection(conn *websocket.Conn) *WSConnection {
	return &WSConnection{conn: conn}
}

func (w *WSConnection) Send(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	error := w.conn.WriteMessage(websocket.TextMessage, data)
	return error
}

func (w *WSConnection) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	_ = w.conn.Close()
}

func (w *WSConnection) Read() ([]byte, error) {

	_, msg, error := w.conn.ReadMessage()

	if error != nil {
		return nil, error
	}
	return msg, nil
}
