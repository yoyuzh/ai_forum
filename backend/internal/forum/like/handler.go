package like

import (
	"context"
	"net/http"
	"strconv"

	"ai-forum/backend/internal/auth"
)

type Actor interface {
	Like(context.Context, DBTX, int64, int64) error
	Unlike(context.Context, DBTX, int64, int64) error
}

type TxRunner func(context.Context, func(DBTX) error) error

type Handler struct {
	service Actor
	runTx   TxRunner
}

func NewHandler(service Actor, runTx TxRunner) *Handler {
	return &Handler{service: service, runTx: runTx}
}

func (h *Handler) Like(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, h.service.Like, "like post")
}

func (h *Handler) Unlike(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, h.service.Unlike, "unlike post")
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request, fn func(context.Context, DBTX, int64, int64) error, msg string) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	postID, err := strconv.ParseInt(r.PathValue("postId"), 10, 64)
	if err != nil || postID <= 0 {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	err = h.runTx(r.Context(), func(tx DBTX) error {
		return fn(r.Context(), tx, sub.UserID, postID)
	})
	if err != nil {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
