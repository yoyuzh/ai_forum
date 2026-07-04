package post

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"ai-forum/backend/internal/auth"
)

type Creator interface {
	CreatePost(context.Context, DBTX, CreateInput) (Post, error)
	List(context.Context, DBTX) ([]Post, error)
	Get(context.Context, DBTX, int64) (Post, error)
	UpdateOwn(context.Context, DBTX, UpdateInput) (Post, error)
	Delete(context.Context, DBTX, int64) error
	UpdateStatus(context.Context, DBTX, int64, string) error
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
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	var created Post
	err := h.runTx(r.Context(), func(tx DBTX) error {
		var err error
		created, err = h.service.CreatePost(r.Context(), tx, CreateInput{AuthorID: sub.UserID, Title: req.Title, Content: req.Content})
		return err
	})
	if err != nil {
		http.Error(w, "create post", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	var posts []Post
	err := h.runTx(r.Context(), func(tx DBTX) error {
		var err error
		posts, err = h.service.List(r.Context(), tx)
		return err
	})
	if err != nil {
		http.Error(w, "list posts", http.StatusInternalServerError)
		return
	}
	if posts == nil {
		posts = []Post{}
	}
	_ = json.NewEncoder(w).Encode(posts)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	postID, ok := postIDFromRequest(w, r)
	if !ok {
		return
	}
	var p Post
	err := h.runTx(r.Context(), func(tx DBTX) error {
		var err error
		p, err = h.service.Get(r.Context(), tx, postID)
		return err
	})
	if err != nil {
		http.Error(w, "get post", http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(p)
}

func (h *Handler) UpdateOwn(w http.ResponseWriter, r *http.Request) {
	sub, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	postID, ok := postIDFromRequest(w, r)
	if !ok {
		return
	}
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	var updated Post
	err := h.runTx(r.Context(), func(tx DBTX) error {
		var err error
		updated, err = h.service.UpdateOwn(r.Context(), tx, UpdateInput{PostID: postID, AuthorID: sub.UserID, Title: req.Title, Content: req.Content})
		return err
	})
	if err != nil {
		http.Error(w, "update post", http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(updated)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.SubjectFromContext(r.Context()); !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	postID, ok := postIDFromRequest(w, r)
	if !ok {
		return
	}
	err := h.runTx(r.Context(), func(tx DBTX) error {
		return h.service.Delete(r.Context(), tx, postID)
	})
	if err != nil {
		http.Error(w, "delete post", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	postID, ok := postIDFromRequest(w, r)
	if !ok {
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	err := h.runTx(r.Context(), func(tx DBTX) error {
		return h.service.UpdateStatus(r.Context(), tx, postID, req.Status)
	})
	if err != nil {
		http.Error(w, "update post status", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func postIDFromRequest(w http.ResponseWriter, r *http.Request) (int64, bool) {
	postID, err := strconv.ParseInt(r.PathValue("postId"), 10, 64)
	if err != nil || postID <= 0 {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return 0, false
	}
	return postID, true
}
