package main

import (
	"context"
	"fmt"
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer stop()

	cfg := configs.Load()

	redis, err := initRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("failed to initialize score repository: %v", err)
	}
	defer redis.Close()

	repo, err := repositories.NewRedisRepository(redis)
	if err != nil {
		log.Fatalf("failed to initialize score repository: %v", err)
	}

	srv := server.New(ctx, cfg.Server.Addr, repo, redis)

	log.Printf("game server listening on %s", cfg.Server.Addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func initRedis(cfg configs.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return client, nil
}
