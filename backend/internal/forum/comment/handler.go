package comment

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"ai-forum/backend/internal/auth"
)

type Creator interface {
	Create(context.Context, DBTX, CreateInput) (Comment, error)
}

type TxRunner func(context.Context, func(DBTX) error) error

type Handler struct {
	service Creator
	runTx   TxRunner
}

func NewHandler(service Creator, runTx TxRunner) *Handler {
	return &Handler{service: service, runTx: runTx}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	postID, err := strconv.ParseInt(r.PathValue("postId"), 10, 64)
	if err != nil || postID <= 0 {
		http.Error(w, "invalid postId", http.StatusBadRequest)
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	var created Comment
	err = h.runTx(r.Context(), func(tx DBTX) error {
		var err error
		created, err = h.service.Create(r.Context(), tx, CreateInput{PostID: postID, UserID: sub.UserID, Content: req.Content})
		return err
	})
	if err != nil {
		http.Error(w, "create comment", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}
