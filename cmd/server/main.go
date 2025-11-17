package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"github.com/sunary/emu-game/configs"
	"github.com/sunary/emu-game/internal/repositories"
	"github.com/sunary/emu-game/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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

	stop := make(chan os.Signal, 1)
	errCh := make(chan error)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("server error: %v", err)
			errCh <- err
		}
	}()

	shutdownFn := func() {
		srv.Shutdown(ctx)
		os.Exit(0)
	}

	for {
		select {
		case <-stop:
			shutdownFn()
			return

		case <-ctx.Done():
			shutdownFn()
			return

		case <-errCh:
			shutdownFn()
			return
		}
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
