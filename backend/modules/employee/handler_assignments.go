package employee

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// assignmentRequest — тело запроса на создание/обновление кадрового
// назначения.
type assignmentRequest struct {
	OrganizationUnitID int64   `json:"organization_unit_id"`
	PositionID         int64   `json:"position_id"`
	AssignmentType     string  `json:"assignment_type"`
	StartsAt           string  `json:"starts_at"`
	EndsAt             *string `json:"ends_at"`
}

func (req assignmentRequest) validate() string {
	if req.OrganizationUnitID <= 0 {
		return "organization_unit_id: обязательное поле"
	}

	if req.PositionID <= 0 {
		return "position_id: обязательное поле"
	}

	switch req.AssignmentType {
	case AssignmentPrimary, AssignmentPartTime, AssignmentTemporaryTransfer, AssignmentCombination:
	default:
		return "assignment_type: недопустимое значение"
	}

	if req.StartsAt == "" {
		return "starts_at: обязательное поле"
	}

	return ""
}

func (req assignmentRequest) toInput() (AssignmentInput, error) {
	startsAt, err := time.Parse(dateLayout, req.StartsAt)
	if err != nil {
		return AssignmentInput{}, err
	}

	in := AssignmentInput{
		OrganizationUnitID: req.OrganizationUnitID,
		PositionID:         req.PositionID,
		AssignmentType:     req.AssignmentType,
		StartsAt:           startsAt,
	}

	if req.EndsAt != nil && *req.EndsAt != "" {
		endsAt, err := time.Parse(dateLayout, *req.EndsAt)
		if err != nil {
			return AssignmentInput{}, err
		}

		in.EndsAt = &endsAt
	}

	return in, nil
}

func (h *Handler) addAssignment(w http.ResponseWriter, r *http.Request) {
	employeeID, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req assignmentRequest
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
		writeError(w, http.StatusBadRequest, "starts_at/ends_at: ожидается формат YYYY-MM-DD")
		return
	}

	a, err := h.repo.AddAssignment(r.Context(), employeeID, in)
	if err != nil {
		writeRepoError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, a)
}

func (h *Handler) updateAssignment(w http.ResponseWriter, r *http.Request) {
	employeeID, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	assignmentID, err := parseID(r.PathValue("assignmentID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid assignment id")
		return
	}

	var req assignmentRequest
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
		writeError(w, http.StatusBadRequest, "starts_at/ends_at: ожидается формат YYYY-MM-DD")
		return
	}

	a, err := h.repo.UpdateAssignment(r.Context(), employeeID, assignmentID, in)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "assignment not found")
			return
		}

		writeRepoError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) deleteAssignment(w http.ResponseWriter, r *http.Request) {
	employeeID, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	assignmentID, err := parseID(r.PathValue("assignmentID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid assignment id")
		return
	}

	if err := h.repo.DeleteAssignment(r.Context(), employeeID, assignmentID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "assignment not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
