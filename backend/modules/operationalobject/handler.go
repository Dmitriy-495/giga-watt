package operationalobject

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// Handler — HTTP-обработчики модуля operationalobject.
type Handler struct {
	repo *Repository
}

// NewHandler создаёт Handler поверх Repository.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes регистрирует маршруты модуля operationalobject в
// мультиплексоре.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/operational-objects", h.list)
	mux.HandleFunc("POST /api/operational-objects", h.create)
	mux.HandleFunc("GET /api/operational-objects/{id}", h.get)
	mux.HandleFunc("PUT /api/operational-objects/{id}", h.update)
	mux.HandleFunc("DELETE /api/operational-objects/{id}", h.delete)
}

// list поддерживает необязательный фильтр
// ?organization_unit_id=, чтобы ограничить список одним ПУ.
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	var organizationUnitID int64

	if raw := r.URL.Query().Get("organization_unit_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "organization_unit_id: invalid value")
			return
		}

		organizationUnitID = id
	}

	objects, err := h.repo.List(r.Context(), organizationUnitID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, objects)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	o, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "operational object not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	writeJSON(w, http.StatusOK, o)
}

// objectRequest — тело запроса на создание/обновление объекта
// эксплуатации.
type objectRequest struct {
	OrganizationUnitID int64    `json:"organization_unit_id"`
	ObjectTypeID       int64    `json:"object_type_id"`
	ObjectPurposeID    *int64   `json:"object_purpose_id"`
	Name               string   `json:"name"`
	Address            *string  `json:"address"`
	Latitude           *float64 `json:"latitude"`
	Longitude          *float64 `json:"longitude"`
}

func (req objectRequest) validate() string {
	if req.OrganizationUnitID <= 0 {
		return "organization_unit_id: обязательное поле"
	}

	if req.ObjectTypeID <= 0 {
		return "object_type_id: обязательное поле"
	}

	if strings.TrimSpace(req.Name) == "" {
		return "name: обязательное поле"
	}

	return ""
}

func (req objectRequest) toInput() Input {
	return Input{
		OrganizationUnitID: req.OrganizationUnitID,
		ObjectTypeID:       req.ObjectTypeID,
		ObjectPurposeID:    req.ObjectPurposeID,
		Name:               req.Name,
		Address:            req.Address,
		Latitude:           req.Latitude,
		Longitude:          req.Longitude,
	}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req objectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	o, err := h.repo.Create(r.Context(), req.toInput())
	if err != nil {
		writeRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, o)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req objectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	o, err := h.repo.Update(r.Context(), id, req.toInput())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "operational object not found")
			return
		}

		writeRepoError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, o)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "operational object not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal error")

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
