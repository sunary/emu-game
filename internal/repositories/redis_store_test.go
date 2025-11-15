package repositories

import (
	"context"
	"fmt"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/sunary/emu-game/internal/models"
)

func newTestRepo(t *testing.T) (*RedisRepository, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)

	repo, err := NewRedisRepository(redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	}))
	require.NoError(t, err)

	return repo, mr
}

func TestRedisRepositoryJoinAndGetQuiz(t *testing.T) {
	repo, mr := newTestRepo(t)
	defer func() {
		mr.Close()
	}()

	ctx := context.Background()

	err := repo.JoinQuiz(ctx, "user-1", "quiz-1")
	require.NoError(t, err)

	quizID, err := repo.GetQuizByUserID(ctx, "user-1")
	require.NoError(t, err)
	require.Equal(t, "quiz-1", quizID)

	empty, err := repo.GetQuizByUserID(ctx, "unknown")
	require.NoError(t, err)
	require.Empty(t, empty)
}

func TestRedisRepositorySubmitAndListScores(t *testing.T) {
	repo, mr := newTestRepo(t)
	defer func() {
		mr.Close()
	}()

	ctx := context.Background()

	err := repo.JoinQuiz(ctx, "user-1", "quiz-1")
	require.NoError(t, err)

	err = repo.SubmitQuiz(ctx, models.UserQuiz{UserID: "user-1", QuizID: "quiz-1", Score: 120})
	require.NoError(t, err)

	err = repo.SubmitQuiz(ctx, models.UserQuiz{UserID: "user-2", QuizID: "quiz-2", Score: 300})
	require.NoError(t, err)

	scores, err := repo.ListUserScores(ctx, 0, 10)
	require.NoError(t, err)
	require.Len(t, scores, 2)
	require.Equal(t, "user-2", scores[0].UserID)
	require.Equal(t, float64(300), scores[0].Score)
	require.Equal(t, "user-1", scores[1].UserID)
	require.Equal(t, float64(120), scores[1].Score)

	quizID, err := repo.GetQuizByUserID(ctx, "user-1")
	require.NoError(t, err)
	require.Empty(t, quizID)
}

func TestRedisRepositoryLeaderboardPagination(t *testing.T) {
	repo, mr := newTestRepo(t)
	defer func() {
		mr.Close()
	}()

	ctx := context.Background()

	for i := 0; i < 15; i++ {
		userID := fmt.Sprintf("user-%02d", i)
		quizID := fmt.Sprintf("quiz-%02d", i)
		score := float64(100 + i)

		require.NoError(t, repo.JoinQuiz(ctx, userID, quizID))
		require.NoError(t, repo.SubmitQuiz(ctx, models.UserQuiz{
			UserID: userID,
			QuizID: quizID,
			Score:  score,
		}))
	}

	list, err := repo.ListUserScores(ctx, 0, 5)
	require.NoError(t, err)
	require.Len(t, list, 5)
	require.Equal(t, "user-14", list[0].UserID)
	require.Equal(t, float64(114), list[0].Score)
	require.Equal(t, "user-10", list[4].UserID)
	require.Equal(t, float64(110), list[4].Score)

	list, err = repo.ListUserScores(ctx, 5, 5)
	require.NoError(t, err)
	require.Len(t, list, 5)
	require.Equal(t, "user-09", list[0].UserID)
	require.Equal(t, float64(109), list[0].Score)
	require.Equal(t, "user-05", list[4].UserID)
	require.Equal(t, float64(105), list[4].Score)
}
