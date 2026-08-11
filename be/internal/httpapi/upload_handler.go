package httpapi

import (
	"net/http"
	"strings"

	"minhquang/be/internal/storage"
)

type UploadHandler struct{ storage *storage.MinIO }

func NewUploadHandler(objectStorage *storage.MinIO) *UploadHandler {
	return &UploadHandler{storage: objectStorage}
}

func (h *UploadHandler) Presign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	if h.storage == nil {
		writeError(w, http.StatusServiceUnavailable, "storage_unavailable", "Kho ảnh chưa được cấu hình")
		return
	}
	var payload struct {
		FileName    string `json:"file_name"`
		ContentType string `json:"content_type"`
		Kind        string `json:"kind"`
	}
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
		return
	}
	if strings.TrimSpace(payload.FileName) == "" || len(payload.FileName) > 255 || payload.Kind != "avatar" || !allowedAvatarType(payload.ContentType) {
		writeError(w, http.StatusBadRequest, "invalid_input", "Chỉ hỗ trợ ảnh JPEG, PNG hoặc WebP")
		return
	}
	result, err := h.storage.PresignAvatar(r.Context(), currentUser(r.Context()).ID, payload.FileName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "presign_failed", "Không thể chuẩn bị tải ảnh")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func allowedAvatarType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}
