package employee

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type contactRequest struct {
	Value     string `json:"phone"`
	Email     string `json:"email"`
	IsPrimary bool   `json:"is_primary"`
}

func (h *Handler) addPhone(w http.ResponseWriter, r *http.Request) {
	employeeID, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req contactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if strings.TrimSpace(req.Value) == "" {
		writeError(w, http.StatusBadRequest, "phone: обязательное поле")
		return
	}

	p, err := h.repo.AddPhone(r.Context(), employeeID, req.Value, req.IsPrimary)
	if err != nil {
		writeRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) updatePhone(w http.ResponseWriter, r *http.Request) {
	employeeID, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	phoneID, err := parseID(r.PathValue("phoneID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid phone id")
		return
	}

	var req contactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if strings.TrimSpace(req.Value) == "" {
		writeError(w, http.StatusBadRequest, "phone: обязательное поле")
		return
	}

	p, err := h.repo.UpdatePhone(r.Context(), employeeID, phoneID, req.Value, req.IsPrimary)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "phone not found")
			return
		}

		writeRepoError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) deletePhone(w http.ResponseWriter, r *http.Request) {
	employeeID, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	phoneID, err := parseID(r.PathValue("phoneID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid phone id")
		return
	}

	if err := h.repo.DeletePhone(r.Context(), employeeID, phoneID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "phone not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) addEmail(w http.ResponseWriter, r *http.Request) {
	employeeID, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req contactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if strings.TrimSpace(req.Email) == "" {
		writeError(w, http.StatusBadRequest, "email: обязательное поле")
		return
	}

	e, err := h.repo.AddEmail(r.Context(), employeeID, req.Email, req.IsPrimary)
	if err != nil {
		writeRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, e)
}

func (h *Handler) updateEmail(w http.ResponseWriter, r *http.Request) {
	employeeID, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	emailID, err := parseID(r.PathValue("emailID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid email id")
		return
	}

	var req contactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if strings.TrimSpace(req.Email) == "" {
		writeError(w, http.StatusBadRequest, "email: обязательное поле")
		return
	}

	e, err := h.repo.UpdateEmail(r.Context(), employeeID, emailID, req.Email, req.IsPrimary)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "email not found")
			return
		}

		writeRepoError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, e)
}

func (h *Handler) deleteEmail(w http.ResponseWriter, r *http.Request) {
	employeeID, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	emailID, err := parseID(r.PathValue("emailID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid email id")
		return
	}

	if err := h.repo.DeleteEmail(r.Context(), employeeID, emailID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "email not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
