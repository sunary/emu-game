package server

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const (
	eventsChannel   = "emu-game:events"
	submitQuizEvent = "submit_quiz_event"
)

type wsHub struct {
	mu     sync.RWMutex
	conns  map[*websocket.Conn]struct{}
	pubsub *redis.PubSub
}

func newHub(redis *redis.Client) *wsHub {
	return &wsHub{
		conns:  make(map[*websocket.Conn]struct{}),
		pubsub: redis.Subscribe(context.Background(), eventsChannel),
	}
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

type eventMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

func (m eventMessage) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func (h *wsHub) subscribe(ctx context.Context) {
	log.Printf("subscribing to events channel")

	ch := h.pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			log.Printf("context done")
			return
		case msg := <-ch:
			log.Printf("received event: %s", msg.Payload)
			var event eventMessage
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("failed to unmarshal event message: %v", err)
				continue
			}

			switch event.Event {
			case submitQuizEvent:
				h.broadcast(event.Data)
			}
		}
	}
}
