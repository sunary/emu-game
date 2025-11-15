package repositories

import (
	"context"

	"github.com/sunary/emu-game/internal/models"
)

type Repository interface {
	JoinQuiz(ctx context.Context, userID string, quizID string) error
	GetQuizByUserID(ctx context.Context, userID string) (string, error)
	SubmitQuiz(ctx context.Context, userQuiz models.UserQuiz) error
	ListUserScores(ctx context.Context, from, limit int64) ([]models.UserQuiz, error)
	Close() error
}
