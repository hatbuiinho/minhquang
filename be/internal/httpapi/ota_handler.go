package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"reminder/be/internal/ota"
)

type OTAHandler struct {
	updates *ota.Service
}

func NewOTAHandler(updates *ota.Service) *OTAHandler {
	return &OTAHandler{updates: updates}
}

func (h *OTAHandler) AndroidLatest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}

	update, err := h.updates.Latest(r.Context(), ota.CheckInput{
		Platform:       "android",
		Channel:        r.URL.Query().Get("channel"),
		CurrentVersion: r.URL.Query().Get("current_version"),
		NativeVersion:  r.URL.Query().Get("native_version"),
	})
	if err != nil {
		if errors.Is(err, ota.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}

	if strings.HasPrefix(update.URL, "/") {
		update.URL = absoluteURL(r, update.URL)
	}

	writeJSON(w, http.StatusOK, update)
}

func absoluteURL(r *http.Request, path string) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwardedProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwardedProto != "" {
		scheme = forwardedProto
	}

	return scheme + "://" + r.Host + path
}
