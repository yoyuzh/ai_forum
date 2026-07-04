package comment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"ai-forum/backend/internal/auth"
)

type Creator interface {
	Create(context.Context, DBTX, CreateInput) (Comment, error)
	List(context.Context, DBTX, int64) ([]Comment, error)
}

type TxRunner func(context.Context, func(DBTX) error) error

type Handler struct {
	service     Creator
	runTx       TxRunner
	afterCommit func(func(context.Context) error)
}

type HandlerOption func(*Handler)

func WithHandlerAfterCommit(after func(func(context.Context) error)) HandlerOption {
	return func(h *Handler) { h.afterCommit = after }
}

func NewHandler(service Creator, runTx TxRunner, opts ...HandlerOption) *Handler {
	h := &Handler{service: service, runTx: runTx}
	for _, opt := range opts {
		opt(h)
	}
	return h
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
		Content         string `json:"content"`
		ParentCommentID *int64 `json:"parent_comment_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	var created Comment
	var callbacks []func(context.Context) error
	txCtx := contextWithAfterCommit(r.Context(), func(fn func(context.Context) error) {
		callbacks = append(callbacks, fn)
	})
	err = h.runTx(txCtx, func(tx DBTX) error {
		var err error
		created, err = h.service.Create(txCtx, tx, CreateInput{PostID: postID, UserID: sub.UserID, ParentCommentID: req.ParentCommentID, Content: req.Content})
		return err
	})
	if err != nil {
		if errors.Is(err, ErrMentionRateLimited) {
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		http.Error(w, "create comment", http.StatusBadRequest)
		return
	}
	if h.afterCommit != nil {
		h.afterCommit(func(ctx context.Context) error {
			for _, fn := range callbacks {
				if err := fn(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	} else {
		for _, fn := range callbacks {
			if err := fn(r.Context()); err != nil {
				http.Error(w, "enqueue comment side effect", http.StatusInternalServerError)
				return
			}
		}
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseInt(r.PathValue("postId"), 10, 64)
	if err != nil || postID <= 0 {
		http.Error(w, "invalid postId", http.StatusBadRequest)
		return
	}
	var comments []Comment
	err = h.runTx(r.Context(), func(tx DBTX) error {
		var err error
		comments, err = h.service.List(r.Context(), tx, postID)
		return err
	})
	if err != nil {
		http.Error(w, "list comments", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(comments)
}
