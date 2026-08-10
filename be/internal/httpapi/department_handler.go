package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"minhquang/be/internal/department"
)

type DepartmentHandler struct{ departments *department.Service }

func NewDepartmentHandler(departments *department.Service) *DepartmentHandler {
	return &DepartmentHandler{departments: departments}
}

func (h *DepartmentHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		active, ok := parseActiveFilter(w, r.URL.Query().Get("active"))
		if !ok {
			return
		}
		items, err := h.departments.List(r.Context(), department.ListOptions{Query: r.URL.Query().Get("q"), Active: active})
		if err != nil {
			handleDepartmentError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string][]department.Department{"departments": items})
	case http.MethodPost:
		var payload struct {
			Name string `json:"name"`
		}
		if err := readJSON(r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
			return
		}
		item, err := h.departments.Create(r.Context(), payload.Name)
		if err != nil {
			handleDepartmentError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func (h *DepartmentHandler) Item(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/departments/"), "/")
	parts := strings.Split(path, "/")
	if path == "" || len(parts) > 2 || (len(parts) == 2 && parts[1] != "status") {
		writeError(w, http.StatusNotFound, "not_found", "Không tìm thấy phân ban")
		return
	}
	id := parts[0]
	if len(parts) == 2 {
		if r.Method != http.MethodPatch {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
			return
		}
		var payload struct {
			Active bool `json:"active"`
		}
		if err := readJSON(r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
			return
		}
		item, err := h.departments.SetActive(r.Context(), id, payload.Active)
		if err != nil {
			handleDepartmentError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}

	switch r.Method {
	case http.MethodPut:
		var payload struct {
			Name string `json:"name"`
		}
		if err := readJSON(r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
			return
		}
		item, err := h.departments.Update(r.Context(), id, payload.Name)
		if err != nil {
			handleDepartmentError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodDelete:
		if err := h.departments.Delete(r.Context(), id); err != nil {
			handleDepartmentError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func (h *DepartmentHandler) Options(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			writeError(w, http.StatusBadRequest, "invalid_input", "limit must be a positive integer")
			return
		}
		limit = parsed
	}
	active := true
	items, err := h.departments.List(r.Context(), department.ListOptions{Query: r.URL.Query().Get("q"), Active: &active, Limit: limit})
	if err != nil {
		handleDepartmentError(w, err)
		return
	}
	names := make([]string, len(items))
	for index, item := range items {
		names[index] = item.Name
	}
	writeJSON(w, http.StatusOK, map[string][]string{"departments": names})
}

func parseActiveFilter(w http.ResponseWriter, value string) (*bool, bool) {
	if value == "" || value == "all" {
		return nil, true
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "active must be true, false, or all")
		return nil, false
	}
	return &parsed, true
}

func handleDepartmentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, department.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, department.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "Không tìm thấy phân ban")
	case errors.Is(err, department.ErrNameExists):
		writeError(w, http.StatusConflict, "name_exists", "Tên phân ban đã tồn tại")
	case errors.Is(err, department.ErrInUse):
		writeError(w, http.StatusConflict, "department_in_use", "Phân ban đang được sử dụng và không thể xoá")
	case errors.Is(err, department.ErrInactive):
		writeError(w, http.StatusConflict, "department_inactive", "Phân ban đã ngừng sử dụng")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
