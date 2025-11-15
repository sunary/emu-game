package server

type joinQuizRequest struct {
}

type submitQuizRequest struct {
	Score float64 `json:"score"`
}

type leaderboardRequest struct {
	From  int64 `json:"from"`
	Limit int64 `json:"limit"`
}
