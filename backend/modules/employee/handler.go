package employee

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

// Handler — HTTP-обработчики модуля employee.
type Handler struct {
	repo *Repository
}

// NewHandler создаёт Handler поверх Repository.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes регистрирует маршруты модуля employee в мультиплексоре.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/employees", h.list)
	mux.HandleFunc("POST /api/employees", h.create)
	mux.HandleFunc("GET /api/employees/{id}", h.get)
	mux.HandleFunc("PUT /api/employees/{id}", h.update)
	mux.HandleFunc("DELETE /api/employees/{id}", h.delete)

	mux.HandleFunc("POST /api/employees/{id}/phones", h.addPhone)
	mux.HandleFunc("PUT /api/employees/{id}/phones/{phoneID}", h.updatePhone)
	mux.HandleFunc("DELETE /api/employees/{id}/phones/{phoneID}", h.deletePhone)

	mux.HandleFunc("POST /api/employees/{id}/emails", h.addEmail)
	mux.HandleFunc("PUT /api/employees/{id}/emails/{emailID}", h.updateEmail)
	mux.HandleFunc("DELETE /api/employees/{id}/emails/{emailID}", h.deleteEmail)

	mux.HandleFunc("POST /api/employees/{id}/assignments", h.addAssignment)
	mux.HandleFunc("PUT /api/employees/{id}/assignments/{assignmentID}", h.updateAssignment)
	mux.HandleFunc("DELETE /api/employees/{id}/assignments/{assignmentID}", h.deleteAssignment)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	employees, err := h.repo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, employees)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	detail, err := h.repo.GetDetail(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	writeJSON(w, http.StatusOK, detail)
}

// employeeRequest — тело запроса на создание/обновление сотрудника.
type employeeRequest struct {
	LastName   string  `json:"last_name"`
	FirstName  string  `json:"first_name"`
	MiddleName string  `json:"middle_name"`
	BirthDate  *string `json:"birth_date"`
	Gender     *string `json:"gender"`
}

func (req employeeRequest) validate() string {
	if strings.TrimSpace(req.LastName) == "" {
		return "last_name: обязательное поле"
	}

	if strings.TrimSpace(req.FirstName) == "" {
		return "first_name: обязательное поле"
	}

	if strings.TrimSpace(req.MiddleName) == "" {
		return "middle_name: обязательное поле"
	}

	if req.Gender != nil {
		switch *req.Gender {
		case GenderMale, GenderFemale:
		default:
			return "gender: недопустимое значение"
		}
	}

	return ""
}

func (req employeeRequest) toInput() (Input, error) {
	in := Input{
		LastName:   req.LastName,
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		Gender:     req.Gender,
	}

	if req.BirthDate != nil && *req.BirthDate != "" {
		d, err := time.Parse(dateLayout, *req.BirthDate)
		if err != nil {
			return Input{}, err
		}

		in.BirthDate = &d
	}

	return in, nil
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req employeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	in, err := req.toInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, "birth_date: ожидается формат YYYY-MM-DD")
		return
	}

	e, err := h.repo.Create(r.Context(), in)
	if err != nil {
		writeRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, e)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req employeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	in, err := req.toInput()
	if err != nil {
		writeError(w, http.StatusBadRequest, "birth_date: ожидается формат YYYY-MM-DD")
		return
	}

	e, err := h.repo.Update(r.Context(), id, in)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}

		writeRepoError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, e)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
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
