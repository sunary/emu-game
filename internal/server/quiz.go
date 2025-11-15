package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sunary/emu-game/internal/models"
	"github.com/sunary/emu-game/pkg"
)

func (a *apiHandlers) joinQuiz(w http.ResponseWriter, r *http.Request) {
	userID := pkg.GetUserID(r.Context())

	var req joinQuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	reqQuizID := mux.Vars(r)["id"]
	if reqQuizID == "" {
		http.Error(w, "quiz ID is required", http.StatusBadRequest)
		return
	}

	quiz, err := a.repo.GetQuizByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("failed to get quiz by user id: %v", err)
		http.Error(w, "failed to get quiz by user id", http.StatusInternalServerError)
		return
	}
	if quiz != "" {
		if quiz == reqQuizID {
			// Protect against accidental rejoin to the same quiz—prevent duplicate state.
			http.Error(w, "user already joined quiz", http.StatusBadRequest)
			return
		}

		// Users cannot be in multiple quizzes simultaneously to avoid conflicting submissions.
		http.Error(w, "user already joined another quiz", http.StatusBadRequest)
		return
	}

	if err := a.repo.JoinQuiz(r.Context(), userID, reqQuizID); err != nil {
		log.Printf("failed to join quiz: %v", err)
		http.Error(w, "failed to join quiz", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": fmt.Sprintf("joined quiz %s", reqQuizID)}); err != nil {
		log.Printf("failed to encode join quiz response: %v", err)
	}
}

func (a *apiHandlers) submitQuiz(w http.ResponseWriter, r *http.Request) {
	userID := pkg.GetUserID(r.Context())

	var req submitQuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	reqQuizID := mux.Vars(r)["id"]
	if reqQuizID == "" {
		http.Error(w, "quiz ID is required", http.StatusBadRequest)
		return
	}

	quiz, err := a.repo.GetQuizByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("failed to get quiz by user id: %v", err)
		http.Error(w, "failed to get quiz by user id", http.StatusInternalServerError)
		return
	}
	if quiz != reqQuizID {
		// Only allow submissions for the quiz the user actually joined.
		http.Error(w, "user did not join quiz", http.StatusBadRequest)
		return
	}

	if err := a.repo.SubmitQuiz(r.Context(), models.UserQuiz{UserID: userID, QuizID: reqQuizID, Score: req.Score}); err != nil {
		log.Printf("failed to submit quiz: %v", err)
		http.Error(w, "failed to submit quiz", http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(models.UserQuiz{UserID: userID, QuizID: reqQuizID, Score: req.Score})
	event := eventMessage{
		Event: submitQuizEvent,
		Data:  data,
	}
	// Publish the event to the Redis channel so that the websocket hub can broadcast it to all connected clients.
	if err := a.redis.Publish(r.Context(), eventsChannel, event).Err(); err != nil {
		log.Printf("failed to publish event: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": fmt.Sprintf("submitted quiz %s", reqQuizID)}); err != nil {
		log.Printf("failed to encode submit quiz response: %v", err)
	}
}

func (a *apiHandlers) leaderboard(w http.ResponseWriter, r *http.Request) {
	var req leaderboardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	scores, err := a.repo.ListUserScores(r.Context(), req.From, req.Limit)
	if err != nil {
		log.Printf("failed to list user scores: %v", err)
		http.Error(w, "failed to list user scores", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(scores); err != nil {
		log.Printf("failed to encode leaderboard response: %v", err)
	}
}
