package httpapi

import (
	"errors"
	"net/http"

	"reminder/be/internal/device"
)

type DeviceHandler struct {
	devices *device.Service
}

type devicePayload struct {
	Platform  device.Platform `json:"platform"`
	PushToken string          `json:"push_token"`
}

func NewDeviceHandler(devices *device.Service) *DeviceHandler {
	return &DeviceHandler{devices: devices}
}

func (h *DeviceHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.register(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func (h *DeviceHandler) register(w http.ResponseWriter, r *http.Request) {
	var payload devicePayload
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
		return
	}

	item, err := h.devices.Register(r.Context(), device.RegisterInput{
		UserID:    userID(r),
		Platform:  payload.Platform,
		PushToken: payload.PushToken,
	})
	if err != nil {
		handleDeviceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func handleDeviceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, device.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
