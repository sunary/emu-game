package models

type UserQuiz struct {
	UserID string  `json:"user_id"`
	QuizID string  `json:"quiz_id"`
	Score  float64 `json:"score"`
}
