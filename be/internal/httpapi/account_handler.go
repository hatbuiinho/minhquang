package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"reminder/be/internal/account"
)

type AccountHandler struct {
	accounts *account.Service
}

type userPayload struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type groupPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewAccountHandler(accounts *account.Service) *AccountHandler {
	return &AccountHandler{accounts: accounts}
}

func (h *AccountHandler) Users(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.accounts.ListUsers(r.Context())
		if err != nil {
			handleAccountError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string][]account.User{"users": items})
	case http.MethodPost:
		var payload userPayload
		if err := readJSON(r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
			return
		}
		item, err := h.accounts.CreateUser(r.Context(), account.CreateUserInput{
			ID:    payload.ID,
			Name:  payload.Name,
			Email: payload.Email,
		})
		if err != nil {
			handleAccountError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func (h *AccountHandler) Groups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.accounts.ListGroups(r.Context())
		if err != nil {
			handleAccountError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string][]account.Group{"groups": items})
	case http.MethodPost:
		var payload groupPayload
		if err := readJSON(r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
			return
		}
		item, err := h.accounts.CreateGroup(r.Context(), account.CreateGroupInput{
			Name:        payload.Name,
			Description: payload.Description,
		})
		if err != nil {
			handleAccountError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func (h *AccountHandler) GroupMembers(w http.ResponseWriter, r *http.Request) {
	groupID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/groups/"), "/members")
	if groupID == "" || strings.Contains(groupID, "/") {
		writeError(w, http.StatusNotFound, "not_found", "group not found")
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}

	var payload struct {
		UserID string `json:"user_id"`
	}
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
		return
	}
	if err := h.accounts.AddGroupMember(r.Context(), groupID, payload.UserID); err != nil {
		handleAccountError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleAccountError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, account.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, account.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "resource not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
