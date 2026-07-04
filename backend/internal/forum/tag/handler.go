package tag

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

type HotTagLister interface {
	ListHotTags(context.Context, DBTX, int) ([]HotTag, error)
}

type TxRunner func(context.Context, func(DBTX) error) error

type Handler struct {
	service HotTagLister
	runTx   TxRunner
}

func NewHandler(service HotTagLister, runTx TxRunner) *Handler {
	return &Handler{service: service, runTx: runTx}
}

func (h *Handler) ListHot(w http.ResponseWriter, r *http.Request) {
	limit := 3
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 20 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	var tags []HotTag
	err := h.runTx(r.Context(), func(tx DBTX) error {
		var err error
		tags, err = h.service.ListHotTags(r.Context(), tx, limit)
		return err
	})
	if err != nil {
		http.Error(w, "list hot tags", http.StatusInternalServerError)
		return
	}
	if tags == nil {
		tags = []HotTag{}
	}
	_ = json.NewEncoder(w).Encode(tags)
}
