package objectpurpose

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// Handler — HTTP-обработчики модуля objectpurpose.
type Handler struct {
	repo *Repository
}

// NewHandler создаёт Handler поверх Repository.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes регистрирует маршруты модуля objectpurpose в
// мультиплексоре.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/object-purposes", h.list)
	mux.HandleFunc("POST /api/object-purposes", h.create)
	mux.HandleFunc("GET /api/object-purposes/{id}", h.get)
	mux.HandleFunc("PUT /api/object-purposes/{id}", h.update)
	mux.HandleFunc("DELETE /api/object-purposes/{id}", h.delete)
}

// list поддерживает необязательный фильтр ?object_type_id=, чтобы
// ограничить список назначений выбранным типом объекта.
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	var objectTypeID int64

	if raw := r.URL.Query().Get("object_type_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "object_type_id: invalid value")
			return
		}

		objectTypeID = id
	}

	purposes, err := h.repo.List(r.Context(), objectTypeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, purposes)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	p, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "object purpose not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	writeJSON(w, http.StatusOK, p)
}

type purposeRequest struct {
	ObjectTypeID int64  `json:"object_type_id"`
	Name         string `json:"name"`
}

func (req purposeRequest) validate() string {
	if req.ObjectTypeID <= 0 {
		return "object_type_id: обязательное поле"
	}

	if strings.TrimSpace(req.Name) == "" {
		return "name: обязательное поле"
	}

	return ""
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req purposeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	p, err := h.repo.Create(r.Context(), req.ObjectTypeID, req.Name)
	if err != nil {
		writeRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req purposeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	p, err := h.repo.Update(r.Context(), id, req.ObjectTypeID, req.Name)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "object purpose not found")
			return
		}

		writeRepoError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "object purpose not found")
			return
		}

		writeRepoError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeRepoError(w http.ResponseWriter, err error) {
	var verr *ValidationError
	if errors.As(err, &verr) {
		writeError(w, http.StatusUnprocessableEntity, verr.Message)
		return
	}

	writeError(w, http.StatusInternalServerError, "internal error")
}

func parseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
