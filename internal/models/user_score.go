package models

import "encoding/json"

type UserQuiz struct {
	UserID string  `json:"user_id"`
	QuizID string  `json:"quiz_id"`
	Score  float64 `json:"score"`
}

func (m UserQuiz) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}
