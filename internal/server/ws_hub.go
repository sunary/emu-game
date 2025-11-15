package server

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type wsHub struct {
	mu    sync.RWMutex
	conns map[*websocket.Conn]struct{}
}

func newHub() *wsHub {
	return &wsHub{conns: make(map[*websocket.Conn]struct{})}
}

func (h *wsHub) add(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[conn] = struct{}{}
}

func (h *wsHub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, conn)
}

func (h *wsHub) broadcast(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.conns {
		conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("failed to send broadcast: %v", err)
			conn.Close()
			go h.remove(conn)
		}
	}
}
