package staffposition

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// Handler — HTTP-обработчики модуля staffposition.
type Handler struct {
	repo *Repository
}

// NewHandler создаёт Handler поверх Repository.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes регистрирует маршруты модуля staffposition в
// мультиплексоре.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/staff-positions", h.list)
	mux.HandleFunc("POST /api/staff-positions", h.create)
	mux.HandleFunc("GET /api/staff-positions/{id}", h.get)
	mux.HandleFunc("PUT /api/staff-positions/{id}", h.update)
	mux.HandleFunc("DELETE /api/staff-positions/{id}", h.delete)
}

// list поддерживает необязательный фильтр
// ?organization_unit_id=, чтобы ограничить список одной
// организационной единицей.
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

	items, err := h.repo.List(r.Context(), organizationUnitID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	s, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "staff position not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	writeJSON(w, http.StatusOK, s)
}

type staffPositionRequest struct {
	OrganizationUnitID int64   `json:"organization_unit_id"`
	PositionID         int64   `json:"position_id"`
	Quantity           float64 `json:"quantity"`
}

func (req staffPositionRequest) validate() string {
	if req.OrganizationUnitID <= 0 {
		return "organization_unit_id: обязательное поле"
	}

	if req.PositionID <= 0 {
		return "position_id: обязательное поле"
	}

	if req.Quantity <= 0 {
		return "quantity: количество единиц должно быть больше нуля"
	}

	return ""
}

func (req staffPositionRequest) toInput() Input {
	return Input{
		OrganizationUnitID: req.OrganizationUnitID,
		PositionID:         req.PositionID,
		Quantity:           req.Quantity,
	}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req staffPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	s, err := h.repo.Create(r.Context(), req.toInput())
	if err != nil {
		writeRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, s)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req staffPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	s, err := h.repo.Update(r.Context(), id, req.toInput())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "staff position not found")
			return
		}

		writeRepoError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, s)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "staff position not found")
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
