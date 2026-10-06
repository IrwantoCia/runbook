package httpapi

import (
	"errors"
	"net/http"

	"runbook/internal/runbook"
)

type publicHandlers struct{ repo runbook.Repository }

func (h publicHandlers) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List("", true, nil)
	if err != nil {
		fail(w, 500, "could not list runbooks")
		return
	}
	jsonResponse(w, 200, map[string]any{"runbooks": items})
}
func (h publicHandlers) get(w http.ResponseWriter, r *http.Request) {
	item, err := h.repo.PublicBySlug(r.PathValue("slug"))
	if errors.Is(err, runbook.ErrNotFound) {
		fail(w, 404, "runbook not found")
		return
	}
	if err != nil {
		fail(w, 500, "could not read runbook")
		return
	}
	jsonResponse(w, 200, map[string]any{"runbook": item})
}
