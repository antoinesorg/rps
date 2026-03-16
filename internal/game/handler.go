package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"rps/internal/logging"
)

// Handler handles HTTP requests for games.
type Handler struct {
	service *Service
}

// NewHandler creates a new game handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateGame handles POST /games.
func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := logging.NewCanonicalLogger(os.Stderr)

	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	logger.Add("action", "create_game")
	logger.Add("request_id", reqID)

	game, err := h.service.CreateGame()
	if err != nil {
		durationMs := float64(time.Since(start).Microseconds()) / 1000.0
		logger.Add("http_status", 500)
		logger.Add("error", err.Error())
		logger.Add("duration_ms", fmt.Sprintf("%.2f", durationMs))
		logger.Emit()

		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to create game.",
		})
		return
	}

	durationMs := float64(time.Since(start).Microseconds()) / 1000.0
	logger.Add("game_id", game.ID)
	logger.Add("http_status", 201)
	logger.Add("duration_ms", fmt.Sprintf("%.2f", durationMs))
	logger.Emit()

	writeJSON(w, http.StatusCreated, game)
}

// PostMove handles POST /games/{id}/moves.
func (h *Handler) PostMove(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := logging.NewCanonicalLogger(os.Stderr)

	id := r.PathValue("id")

	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	logger.Add("action", "make_move")
	logger.Add("request_id", reqID)
	logger.Add("game_id", id)

	var body struct {
		Player interface{} `json:"player"`
		Choice interface{} `json:"choice"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		body.Player = nil
		body.Choice = nil
	}

	// Extract player as int
	var player int
	switch v := body.Player.(type) {
	case float64:
		player = int(v)
	default:
		player = 0
	}

	// Extract choice as string
	var choice string
	switch v := body.Choice.(type) {
	case string:
		choice = v
	default:
		choice = ""
	}

	logger.Add("player", body.Player)
	logger.Add("choice", body.Choice)

	game, err := h.service.MakeMove(id, player, choice)
	if err != nil {
		durationMs := float64(time.Since(start).Microseconds()) / 1000.0

		if errors.Is(err, ErrInvalidArgument) {
			// Extract the message after the wrapped error
			msg := err.Error()
			msg = strings.TrimPrefix(msg, "invalid argument: ")

			logger.Add("http_status", 422)
			logger.Add("error", msg)
			logger.Add("duration_ms", fmt.Sprintf("%.2f", durationMs))
			logger.Emit()

			writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
				"error": msg,
			})
			return
		}

		if errors.Is(err, ErrNotFound) {
			logger.Add("http_status", 404)
			logger.Add("error", "Game not found.")
			logger.Add("duration_ms", fmt.Sprintf("%.2f", durationMs))
			logger.Emit()

			writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"error": "Game not found.",
			})
			return
		}

		logger.Add("http_status", 500)
		logger.Add("error", err.Error())
		logger.Add("duration_ms", fmt.Sprintf("%.2f", durationMs))
		logger.Emit()

		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Unexpected error.",
		})
		return
	}

	durationMs := float64(time.Since(start).Microseconds()) / 1000.0
	logger.Add("game_status", game.Status)
	logger.Add("winner", game.Winner)
	logger.Add("http_status", 200)
	logger.Add("duration_ms", fmt.Sprintf("%.2f", durationMs))
	logger.Emit()

	writeJSON(w, http.StatusOK, game)
}

// HealthCheck handles GET /games for health check purposes.
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
