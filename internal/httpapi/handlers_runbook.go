package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"runbook/internal/runbook"
)

type runbookHandlers struct {
	repo    runbook.Repository
	service runbook.Service
}

func (h runbookHandlers) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List(r.URL.Query().Get("status"), false, optionalQuery(r, "folder_id"))
	if err != nil {
		fail(w, 500, "could not list runbooks")
		return
	}
	jsonResponse(w, 200, map[string]any{"runbooks": items})
}
func (h runbookHandlers) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title    string  `json:"title"`
		Body     string  `json:"body"`
		FolderID *string `json:"folder_id"`
	}
	if decode(r, &input) != nil || strings.TrimSpace(input.Title) == "" {
		fail(w, 400, "title and body are required")
		return
	}
	item, err := h.service.Create(input.Title, input.Body, currentUser(r).ID, input.FolderID)
	if err != nil {
		fail(w, 500, "could not create runbook")
		return
	}
	jsonResponse(w, 201, map[string]any{"runbook": item})
}

func (h runbookHandlers) move(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FolderID *string `json:"folder_id"`
	}
	if decode(r, &input) != nil {
		fail(w, 400, "invalid folder_id")
		return
	}
	if input.FolderID != nil {
		var exists int
		if err := h.repo.DB.QueryRow(`SELECT 1 FROM folders WHERE id=?`, *input.FolderID).Scan(&exists); err == sql.ErrNoRows {
			fail(w, 404, "folder not found")
			return
		} else if err != nil {
			fail(w, 500, "could not validate folder")
			return
		}
	}
	item, err := h.repo.Move(r.PathValue("id"), input.FolderID)
	if errors.Is(err, runbook.ErrNotFound) {
		fail(w, 404, "runbook not found")
		return
	}
	if err != nil {
		fail(w, 500, "could not move runbook")
		return
	}
	jsonResponse(w, 200, map[string]any{"runbook": item})
}

func optionalQuery(r *http.Request, key string) *string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil
	}
	return &value
}
func (h runbookHandlers) get(w http.ResponseWriter, r *http.Request) {
	item, err := h.repo.Get(r.PathValue("id"))
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
func (h runbookHandlers) update(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title *string `json:"title"`
		Body  *string `json:"body"`
	}
	if decode(r, &input) != nil || input.Title == nil && input.Body == nil {
		fail(w, 400, "provide title or body")
		return
	}
	item, err := h.repo.Update(r.PathValue("id"), input.Title, input.Body)
	if errors.Is(err, runbook.ErrNotFound) {
		fail(w, 404, "runbook not found")
		return
	}
	if err != nil {
		fail(w, 409, err.Error())
		return
	}
	jsonResponse(w, 200, map[string]any{"runbook": item})
}
func (h runbookHandlers) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Delete(r.PathValue("id")); err != nil {
		status := 500
		if errors.Is(err, runbook.ErrNotFound) {
			status = 404
		}
		fail(w, status, "could not delete runbook")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h runbookHandlers) publish(w http.ResponseWriter, r *http.Request) {
	err := h.service.Publish(r.PathValue("id"))
		if errors.Is(err, runbook.ErrNotFound) {
			fail(w, 404, "runbook not found")
			return
		}
		if err != nil {
			fail(w, 409, err.Error())
			return
		}
		item, err := h.repo.Get(r.PathValue("id"))
		if err != nil {
			fail(w, 500, "could not read runbook")
			return
		}
		jsonResponse(w, 200, map[string]any{"runbook": item})
}
