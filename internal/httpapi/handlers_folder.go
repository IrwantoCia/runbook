package httpapi

import (
	"errors"
	"net/http"

	"runbook/internal/folder"
)

type folderHandlers struct{ repo folder.Repository }

func (h folderHandlers) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List()
	if err != nil {
		fail(w, 500, "could not list folders")
		return
	}
	jsonResponse(w, 200, map[string]any{"folders": items})
}
func (h folderHandlers) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string  `json:"name"`
		ParentID *string `json:"parent_id"`
	}
	if decode(r, &input) != nil {
		fail(w, 400, "invalid folder")
		return
	}
	id, err := newFolderID()
	if err != nil {
		fail(w, 500, "could not create folder")
		return
	}
	item, err := h.repo.Create(id, input.Name, input.ParentID)
	if err != nil {
		h.folderError(w, err)
		return
	}
	jsonResponse(w, 201, map[string]any{"folder": item})
}
func (h folderHandlers) rename(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if decode(r, &input) != nil {
		fail(w, 400, "invalid folder name")
		return
	}
	item, err := h.repo.Rename(r.PathValue("id"), input.Name)
	if err != nil {
		h.folderError(w, err)
		return
	}
	jsonResponse(w, 200, map[string]any{"folder": item})
}
func (h folderHandlers) move(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ParentID *string `json:"parent_id"`
	}
	if decode(r, &input) != nil {
		fail(w, 400, "invalid parent_id")
		return
	}
	if err := h.repo.Move(r.PathValue("id"), input.ParentID); err != nil {
		h.folderError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h folderHandlers) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Delete(r.PathValue("id")); err != nil {
		h.folderError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h folderHandlers) folderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, folder.ErrNotFound):
		fail(w, 404, err.Error())
	case errors.Is(err, folder.ErrInvalid):
		fail(w, 400, err.Error())
	case errors.Is(err, folder.ErrNotEmpty), errors.Is(err, folder.ErrCycle), errors.Is(err, folder.ErrDuplicate):
		fail(w, 409, err.Error())
	default:
		fail(w, 500, "could not update folder")
	}
}
