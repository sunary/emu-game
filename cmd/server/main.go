package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"github.com/sunary/emu-game/configs"
	"github.com/sunary/emu-game/internal/repositories"
	"github.com/sunary/emu-game/internal/server"
)

func main() {
	// Note: SIGKILL cannot be intercepted by userland programs, but we include it for completeness—
	// the runtime simply never delivers it.
	_, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer stop()

	cfg := configs.Load()

	repo, err := initScoreRepository(cfg.Redis)
	if err != nil {
		log.Fatalf("failed to initialize score repository: %v", err)
	}
	defer repo.Close()

	srv := server.New(cfg.Server.Addr, repo)

	log.Printf("game server listening on %s", cfg.Server.Addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func initScoreRepository(cfg configs.RedisConfig) (repositories.Repository, error) {
	return repositories.NewRedisRepository(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}
