package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"ai-forum/backend/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	u, err := h.service.Register(r.Context(), req)
	if errors.Is(err, ErrDuplicateUsername) {
		http.Error(w, "duplicate username", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "register user", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"id": u.ID, "username": u.Username})
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	u, err := h.service.Profile(r.Context(), sub.UserID)
	if err != nil {
		http.Error(w, "profile", http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":           u.ID,
		"username":     u.Username,
		"email":        u.Email,
		"display_name": u.DisplayName,
		"role":         u.Role,
		"status":       u.Status,
	})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		DisplayName string `json:"display_name"`
		Nickname    string `json:"nickname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Nickname
	}
	u, err := h.service.UpdateProfile(r.Context(), sub.UserID, UpdateProfileInput{DisplayName: displayName})
	if err != nil {
		http.Error(w, "update profile", http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":           u.ID,
		"username":     u.Username,
		"email":        u.Email,
		"display_name": u.DisplayName,
		"role":         u.Role,
		"status":       u.Status,
	})
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	stats, err := h.service.Stats(r.Context(), sub.UserID)
	if err != nil {
		http.Error(w, "profile stats", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(stats)
}
