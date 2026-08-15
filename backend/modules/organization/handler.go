package organization

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// Handler — HTTP-обработчики модуля organization.
type Handler struct {
	repo *Repository
}

// NewHandler создаёт Handler поверх Repository.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes регистрирует маршруты модуля organization в мультиплексоре.
//
// Намеренно нет DELETE: организационная единица — структурный костяк
// (см. docs/iterations/001-foundation/NOTES.md, Decision Log,
// 2026-08-15), на который могут ссылаться сотрудники, штатные единицы и
// объекты эксплуатации. Удаление без согласованного правила о судьбе
// зависимых данных рискует осиротить их или молча потерять
// организационную историю.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/organization-units", h.list)
	mux.HandleFunc("POST /api/organization-units", h.create)
	mux.HandleFunc("GET /api/organization-units/{id}", h.get)
	mux.HandleFunc("PUT /api/organization-units/{id}", h.update)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	units, err := h.repo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, units)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	unit, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "organization unit not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	writeJSON(w, http.StatusOK, unit)
}

// unitRequest — тело запроса на создание/обновление организационной
// единицы.
type unitRequest struct {
	ParentID         *int64   `json:"parent_id"`
	Type             string   `json:"type"`
	Name             string   `json:"name"`
	Location         string   `json:"location"`
	Address          string   `json:"address"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	Phone            *string  `json:"phone"`
	Email            *string  `json:"email"`
	LeaderEmployeeID *int64   `json:"leader_employee_id"`
}

func (req unitRequest) toInput() Input {
	return Input{
		ParentID:         req.ParentID,
		Type:             req.Type,
		Name:             req.Name,
		Location:         req.Location,
		Address:          req.Address,
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		Phone:            req.Phone,
		Email:            req.Email,
		LeaderEmployeeID: req.LeaderEmployeeID,
	}
}

// validate проверяет обязательные поля, которые не завязаны на состояние
// БД (тип, наименование, местоположение, адрес). Проверка иерархии
// (соответствие parent_id и type) остаётся на стороне БД-триггера.
func (req unitRequest) validate() string {
	switch req.Type {
	case TypeInstitution, TypeBranch, TypeJKS, TypeProductionUnit:
	default:
		return "type: недопустимое значение"
	}

	if req.Name == "" {
		return "name: обязательное поле"
	}

	if req.Location == "" {
		return "location: обязательное поле"
	}

	if req.Address == "" {
		return "address: обязательное поле"
	}

	return ""
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req unitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	unit, err := h.repo.Create(r.Context(), req.toInput())
	if err != nil {
		writeRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, unit)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req unitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	unit, err := h.repo.Update(r.Context(), id, req.toInput())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "organization unit not found")
			return
		}

		writeRepoError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, unit)
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
