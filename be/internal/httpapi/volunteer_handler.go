package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"minhquang/be/internal/volunteer"
)

type VolunteerHandler struct{ volunteers *volunteer.Service }

type volunteerPayload struct {
	FullName         string `json:"full_name"`
	DharmaName       string `json:"dharma_name"`
	BirthDate        string `json:"birth_date"`
	CultivationPlace string `json:"cultivation_place"`
	Phone            string `json:"phone"`
	Department       string `json:"department"`
	Notes            string `json:"notes"`
	AvatarURL        string `json:"avatar_url"`
	ArrivalDate      string `json:"arrival_date"`
	DepartureDate    string `json:"departure_date"`
}

func NewVolunteerHandler(volunteers *volunteer.Service) *VolunteerHandler {
	return &VolunteerHandler{volunteers: volunteers}
}

func (h *VolunteerHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.volunteers.List(r.Context(), volunteer.ListOptions{Query: r.URL.Query().Get("q"), Status: r.URL.Query().Get("status")})
		if err != nil {
			handleVolunteerError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string][]volunteer.Volunteer{"volunteers": items})
	case http.MethodPost:
		input, ok := readVolunteerInput(w, r)
		if !ok {
			return
		}
		item, err := h.volunteers.Create(r.Context(), input)
		if err != nil {
			handleVolunteerError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func (h *VolunteerHandler) Item(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/volunteers/"), "/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "not_found", "Không tìm thấy hồ sơ")
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := h.volunteers.Get(r.Context(), id)
		if err != nil {
			handleVolunteerError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPut:
		input, ok := readVolunteerInput(w, r)
		if !ok {
			return
		}
		item, err := h.volunteers.Update(r.Context(), id, input)
		if err != nil {
			handleVolunteerError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodDelete:
		if err := h.volunteers.Delete(r.Context(), id); err != nil {
			handleVolunteerError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func readVolunteerInput(w http.ResponseWriter, r *http.Request) (volunteer.Input, bool) {
	var payload volunteerPayload
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
		return volunteer.Input{}, false
	}
	arrival, err := time.Parse("2006-01-02", payload.ArrivalDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "Ngày đến không hợp lệ")
		return volunteer.Input{}, false
	}
	var departure *time.Time
	if payload.DepartureDate != "" {
		value, err := time.Parse("2006-01-02", payload.DepartureDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "Ngày ra về không hợp lệ")
			return volunteer.Input{}, false
		}
		departure = &value
	}
	return volunteer.Input{FullName: payload.FullName, DharmaName: payload.DharmaName, BirthDate: payload.BirthDate, CultivationPlace: payload.CultivationPlace, Phone: payload.Phone, Department: payload.Department, Notes: payload.Notes, AvatarURL: payload.AvatarURL, ArrivalDate: arrival, DepartureDate: departure}, true
}

func handleVolunteerError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, volunteer.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, volunteer.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "Không tìm thấy hồ sơ")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
