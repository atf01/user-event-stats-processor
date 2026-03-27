package api

import (
	"encoding/json"
	"net/http"
	"user-event-stats-processor/internal/models"
	"user-event-stats-processor/internal/processor"
	"user-event-stats-processor/internal/store"
)

type Handler struct {
	store   store.Storer
	workers *processor.WorkerPool
}

func NewHandler(s store.Storer, wp *processor.WorkerPool) *Handler {
	return &Handler{store: s, workers: wp}
}

func (h *Handler) PostEvent(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var e models.Event
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if e.UserID == "" || e.EventType == "" {
		http.Error(w, "Missing user_id or event_type", http.StatusBadRequest)
		return
	}

	if !h.workers.Enqueue(e) {
		http.Error(w, "Server busy", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	stats := h.store.GetStats(userID)

	if stats == nil {
		stats = make(map[string]float64)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
