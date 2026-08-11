package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"minhquang/be/internal/user"
)

type authContextKey struct{}

type AuthHandler struct{ users *user.Service }

func NewAuthHandler(users *user.Service) *AuthHandler { return &AuthHandler{users: users} }

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
		return
	}
	result, err := h.users.Login(r.Context(), payload.Username, payload.Password)
	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "Tên đăng nhập hoặc mật khẩu không đúng")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	_ = h.users.Logout(r.Context(), bearerToken(r))
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	writeJSON(w, http.StatusOK, currentUser(r.Context()))
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	var payload struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
		return
	}
	err := h.users.ChangePassword(r.Context(), currentUser(r.Context()), payload.CurrentPassword, payload.NewPassword, bearerToken(r))
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, user.ErrCurrentPassword):
		writeError(w, http.StatusBadRequest, "current_password_incorrect", "Mật khẩu hiện tại không đúng")
	case errors.Is(err, user.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	var payload struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	}
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
		return
	}
	item, err := h.users.UpdateProfile(r.Context(), currentUser(r.Context()), payload.Username, payload.DisplayName)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, item)
	case errors.Is(err, user.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, user.ErrUsernameExists):
		writeError(w, http.StatusConflict, "username_exists", "Tên đăng nhập đã tồn tại")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func (h *AuthHandler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	var payload struct {
		AvatarURL string `json:"avatar_url"`
	}
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
		return
	}
	item, err := h.users.UpdateAvatar(r.Context(), currentUser(r.Context()), payload.AvatarURL)
	if err == nil {
		writeJSON(w, http.StatusOK, item)
		return
	}
	if errors.Is(err, user.ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
}

func requireAuth(users *user.Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		item, err := users.Authenticate(r.Context(), bearerToken(r))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Vui lòng đăng nhập")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authContextKey{}, item)))
	})
}

func bearerToken(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(value, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
}

func currentUser(ctx context.Context) user.User {
	item, _ := ctx.Value(authContextKey{}).(user.User)
	return item
}
