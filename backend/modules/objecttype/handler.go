package objecttype

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// Handler — HTTP-обработчики модуля objecttype.
type Handler struct {
	repo *Repository
}

// NewHandler создаёт Handler поверх Repository.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes регистрирует маршруты модуля objecttype в мультиплексоре.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/object-types", h.list)
	mux.HandleFunc("POST /api/object-types", h.create)
	mux.HandleFunc("GET /api/object-types/{id}", h.get)
	mux.HandleFunc("PUT /api/object-types/{id}", h.update)
	mux.HandleFunc("DELETE /api/object-types/{id}", h.delete)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	types, err := h.repo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, types)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	t, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "object type not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	writeJSON(w, http.StatusOK, t)
}

type typeRequest struct {
	Name string `json:"name"`
}

func (req typeRequest) validate() string {
	if strings.TrimSpace(req.Name) == "" {
		return "name: обязательное поле"
	}

	return ""
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req typeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	t, err := h.repo.Create(r.Context(), req.Name)
	if err != nil {
		writeRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req typeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	t, err := h.repo.Update(r.Context(), id, req.Name)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "object type not found")
			return
		}

		writeRepoError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "object type not found")
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
