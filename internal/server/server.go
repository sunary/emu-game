package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/sunary/emu-game/internal/external"
	"github.com/sunary/emu-game/internal/repositories"
	"github.com/sunary/emu-game/pkg"
)

const (
	wsWriteWait  = 10 * time.Second
	wsPongWait   = 60 * time.Second
	wsPingPeriod = (wsPongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type healthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

type apiHandlers struct {
	repo repositories.Repository
	hub  *wsHub
}

func New(addr string, repo repositories.Repository) *http.Server {
	router := mux.NewRouter()
	api := &apiHandlers{repo: repo, hub: newHub()}

	router.Use(userAuthMiddleware())
	router.HandleFunc("/health", healthHandler).Methods(http.MethodGet)
	router.HandleFunc("/ws", wsHandler(api.hub)).Methods(http.MethodGet)
	router.HandleFunc("/user/quiz/{id}/join", api.joinQuiz).Methods(http.MethodPost)
	router.HandleFunc("/user/quiz/{id}/submit", api.submitQuiz).Methods(http.MethodPost)
	router.HandleFunc("/leaderboard", api.leaderboard).Methods(http.MethodGet)

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("request: %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	return &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := healthResponse{Status: "ok", Time: time.Now().UTC().Format(time.RFC3339)}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write health response: %v", err)
	}
}

func wsHandler(hub *wsHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("failed to upgrade connection: %v", err)
			return
		}

		// Heartbeat configuration: enforce read limits to prevent memory pressure, and
		// refresh deadlines whenever we receive a pong so that idle clients are detected.
		conn.SetReadLimit(1024)
		conn.SetReadDeadline(time.Now().Add(wsPongWait))
		conn.SetPongHandler(func(appData string) error {
			conn.SetReadDeadline(time.Now().Add(wsPongWait))
			return nil
		})

		hub.add(conn)
		done := make(chan struct{})
		// Ping loop ensures clients stay responsive; if a ping write fails the connection is closed.
		go func() {
			ticker := time.NewTicker(wsPingPeriod)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
					if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						log.Printf("ping error: %v", err)
						conn.Close()
						return
					}
				case <-done:
					return
				}
			}
		}()

		defer func() {
			close(done)
			hub.remove(conn)
			conn.Close()
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("read error: %v", err)
				return
			}

			log.Printf("received message: %s", string(message))
		}
	}
}

func userAuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/user") {
				next.ServeHTTP(w, r)
				return
			}

			payload, err := external.ValidateJWT(r.Header.Get("Authorization"))
			if err != nil {
				status := http.StatusUnauthorized
				log.Printf("jwt validation failed: %v", err)
				http.Error(w, http.StatusText(status), status)
				return
			}

			ctx := pkg.WithUserID(r.Context(), payload.Sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
