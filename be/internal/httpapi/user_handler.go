package httpapi

import (
	"errors"
	"net/http"

	"minhquang/be/internal/user"
)

type UserHandler struct{ users *user.Service }

func NewUserHandler(users *user.Service) *UserHandler { return &UserHandler{users: users} }

func (h *UserHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.users.List(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			return
		}
		writeJSON(w, http.StatusOK, map[string][]user.User{"users": items})
	case http.MethodPost:
		var payload struct {
			Username    string `json:"username"`
			DisplayName string `json:"display_name"`
			Password    string `json:"password"`
			Role        string `json:"role"`
		}
		if err := readJSON(r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
			return
		}
		item, err := h.users.Create(r.Context(), user.CreateInput{Username: payload.Username, DisplayName: payload.DisplayName, Password: payload.Password, Role: payload.Role})
		if err != nil {
			switch {
			case errors.Is(err, user.ErrInvalidInput):
				writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
			case errors.Is(err, user.ErrUsernameExists):
				writeError(w, http.StatusConflict, "username_exists", "Tên đăng nhập đã tồn tại")
			default:
				writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			}
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func (h *UserHandler) Item(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	var payload struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
		Password    string `json:"password"`
	}
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
		return
	}
	item, err := h.users.Update(r.Context(), currentUser(r.Context()), r.PathValue("id"), user.UpdateInput{
		Username: payload.Username, DisplayName: payload.DisplayName, Role: payload.Role, Password: payload.Password,
	})
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, item)
	case errors.Is(err, user.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, user.ErrUsernameExists):
		writeError(w, http.StatusConflict, "username_exists", "Tên đăng nhập đã tồn tại")
	case errors.Is(err, user.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "Không tìm thấy tài khoản")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
