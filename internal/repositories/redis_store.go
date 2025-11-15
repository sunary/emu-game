package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sunary/emu-game/internal/models"
)

const (
	sortedSetKey  = "emu-game:scores"
	userQuizKeyNS = "emu-game:user"
	expireTime    = 1 * time.Hour
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(redis *redis.Client) (*RedisRepository, error) {
	return &RedisRepository{client: redis}, nil
}

func (s *RedisRepository) JoinQuiz(ctx context.Context, userID, quizID string) error {
	// Store quiz membership with an expiration so abandoned sessions eventually clear.
	return s.client.Set(ctx, userQuizKey(userID), quizID, expireTime).Err()
}

func (s *RedisRepository) GetQuizByUserID(ctx context.Context, userID string) (string, error) {
	val, err := s.client.Get(ctx, userQuizKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}

	return val, nil
}

func (s *RedisRepository) SubmitQuiz(ctx context.Context, userQuiz models.UserQuiz) error {
	payload, err := json.Marshal(userQuiz)
	if err != nil {
		return err
	}

	pipe := s.client.TxPipeline()
	// Remove membership so the user must explicitly re-join before another submit.
	pipe.Del(ctx, userQuizKey(userQuiz.UserID))
	pipe.ZAdd(ctx, sortedSetKey, redis.Z{
		Member: payload,
		Score:  userQuiz.Score,
	})

	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisRepository) ListUserScores(ctx context.Context, from, limit int64) ([]models.UserQuiz, error) {
	if from < 0 {
		from = 0
	}

	if limit <= 0 {
		limit = 10
	}

	vals, err := s.client.ZRevRangeWithScores(ctx, sortedSetKey, from, from+limit-1).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]models.UserQuiz, 0, len(vals))
	for _, v := range vals {
		var quiz models.UserQuiz

		switch raw := v.Member.(type) {
		case string:
			if err := json.Unmarshal([]byte(raw), &quiz); err != nil {
				continue
			}
		case []byte:
			if err := json.Unmarshal(raw, &quiz); err != nil {
				continue
			}
		default:
			continue
		}

		quiz.Score = v.Score
		entries = append(entries, quiz)
	}

	return entries, nil
}

func userQuizKey(userID string) string {
	return fmt.Sprintf("%s:%s", userQuizKeyNS, userID)
}
