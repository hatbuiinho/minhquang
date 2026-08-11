package httpapi

import (
	"net/http"

	"minhquang/be/internal/user"
)

func requirePermission(permission user.Permission, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !user.HasPermission(currentUser(r.Context()).Role, permission) {
			writeError(w, http.StatusForbidden, "forbidden", "Bạn không có quyền thực hiện thao tác này")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requireMethodPermissions(permissions map[string]user.Permission, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		permission, protected := permissions[r.Method]
		if protected && !user.HasPermission(currentUser(r.Context()).Role, permission) {
			writeError(w, http.StatusForbidden, "forbidden", "Bạn không có quyền thực hiện thao tác này")
			return
		}
		next.ServeHTTP(w, r)
	})
}
