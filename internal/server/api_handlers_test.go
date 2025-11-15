package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/sunary/emu-game/internal/models"
	"github.com/sunary/emu-game/internal/repositories"
	"github.com/sunary/emu-game/pkg"
)

type mockRepository struct {
	joinArgs struct {
		ctx    context.Context
		userID string
		quizID string
	}
	joinErr error

	getQuizResult string
	getQuizErr    error

	submitArgs []models.UserQuiz
	submitErr  error

	listArgs struct {
		ctx  context.Context
		from int64
		lim  int64
	}
	listResult []models.UserQuiz
	listErr    error
}

func (m *mockRepository) JoinQuiz(ctx context.Context, userID string, quizID string) error {
	m.joinArgs.ctx = ctx
	m.joinArgs.userID = userID
	m.joinArgs.quizID = quizID
	return m.joinErr
}

func (m *mockRepository) GetQuizByUserID(ctx context.Context, userID string) (string, error) {
	return m.getQuizResult, m.getQuizErr
}

func (m *mockRepository) SubmitQuiz(ctx context.Context, quiz models.UserQuiz) error {
	m.submitArgs = append(m.submitArgs, quiz)
	return m.submitErr
}

func (m *mockRepository) ListUserScores(ctx context.Context, from, limit int64) ([]models.UserQuiz, error) {
	m.listArgs.ctx = ctx
	m.listArgs.from = from
	m.listArgs.lim = limit
	if m.listResult == nil {
		return nil, m.listErr
	}
	return m.listResult, m.listErr
}

func (m *mockRepository) Close() error { return nil }

func newAPIHandlers(t *testing.T, repo repositories.Repository) *apiHandlers {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer mr.Close()
	return &apiHandlers{
		repo:  repo,
		hub:   newHub(redisClient),
		redis: redisClient,
	}
}

func withUserContext(req *http.Request, userID string) *http.Request {
	ctx := pkg.WithUserID(req.Context(), userID)
	return req.WithContext(ctx)
}

func TestJoinQuiz_Success(t *testing.T) {
	repo := &mockRepository{}
	api := newAPIHandlers(t, repo)

	body, _ := json.Marshal(map[string]string{"quiz_id": "quiz-42"})
	req := httptest.NewRequest(http.MethodPost, "/user/quiz-42/join", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "quiz-42"})
	req = withUserContext(req, "user-123")

	rec := httptest.NewRecorder()
	api.joinQuiz(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, "user-123", repo.joinArgs.userID)
	require.Equal(t, "quiz-42", repo.joinArgs.quizID)
}

func TestJoinQuiz_AlreadyJoined(t *testing.T) {
	repo := &mockRepository{getQuizResult: "quiz-42"}
	api := newAPIHandlers(t, repo)

	req := httptest.NewRequest(http.MethodPost, "/user/quiz-42/join", bytes.NewBufferString(`{}`))
	req = mux.SetURLVars(req, map[string]string{"id": "quiz-42"})
	req = withUserContext(req, "user-123")

	rec := httptest.NewRecorder()
	api.joinQuiz(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Empty(t, repo.joinArgs.userID)
}

func TestSubmitQuiz_Success(t *testing.T) {
	repo := &mockRepository{
		getQuizResult: "quiz-99",
	}
	api := newAPIHandlers(t, repo)

	req := httptest.NewRequest(http.MethodPost, "/user/quiz-99/submit", bytes.NewBufferString(`{"score":75}`))
	req = mux.SetURLVars(req, map[string]string{"id": "quiz-99"})
	req = withUserContext(req, "user-abc")

	rec := httptest.NewRecorder()
	api.submitQuiz(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, repo.submitArgs, 1)
	require.Equal(t, models.UserQuiz{UserID: "user-abc", QuizID: "quiz-99", Score: 75}, repo.submitArgs[0])
}

func TestLeaderboard_ReturnsScores(t *testing.T) {
	repo := &mockRepository{
		listResult: []models.UserQuiz{
			{UserID: "u1", QuizID: "q1", Score: 100},
		},
	}
	api := newAPIHandlers(t, repo)

	req := httptest.NewRequest(http.MethodGet, "/leaderboard", bytes.NewBufferString(`{"from":0,"limit":5}`))
	rec := httptest.NewRecorder()

	api.leaderboard(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `[{"user_id":"u1","quiz_id":"q1","score":100}]`, rec.Body.String())
	require.Equal(t, int64(0), repo.listArgs.from)
	require.Equal(t, int64(5), repo.listArgs.lim)
}
