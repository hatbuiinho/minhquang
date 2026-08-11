package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"minhquang/be/internal/user"
)

func TestRequirePermission(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := requirePermission(user.PermissionUserManage, next)

	for _, test := range []struct {
		role string
		want int
	}{{user.RoleAdmin, http.StatusNoContent}, {user.RoleViewer, http.StatusForbidden}} {
		request := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		request = request.WithContext(context.WithValue(request.Context(), authContextKey{}, user.User{Role: test.role}))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != test.want {
			t.Errorf("role %s: expected %d, got %d", test.role, test.want, response.Code)
		}
	}
}

func TestRequireMethodPermissionsAllowsViewerReads(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := requireMethodPermissions(map[string]user.Permission{http.MethodPost: user.PermissionVolunteerCreate}, next)

	for _, test := range []struct {
		method string
		want   int
	}{{http.MethodGet, http.StatusNoContent}, {http.MethodPost, http.StatusForbidden}} {
		request := httptest.NewRequest(test.method, "/api/volunteers", nil)
		request = request.WithContext(context.WithValue(request.Context(), authContextKey{}, user.User{Role: user.RoleViewer}))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != test.want {
			t.Errorf("method %s: expected %d, got %d", test.method, test.want, response.Code)
		}
	}
}

func TestUserPermissionsAllowViewerListOnly(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := requireMethodPermissions(map[string]user.Permission{
		http.MethodGet:  user.PermissionUserRead,
		http.MethodPost: user.PermissionUserManage,
	}, next)

	for _, test := range []struct {
		method string
		want   int
	}{{http.MethodGet, http.StatusNoContent}, {http.MethodPost, http.StatusForbidden}} {
		request := httptest.NewRequest(test.method, "/api/users", nil)
		request = request.WithContext(context.WithValue(request.Context(), authContextKey{}, user.User{Role: user.RoleViewer}))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != test.want {
			t.Errorf("method %s: expected %d, got %d", test.method, test.want, response.Code)
		}
	}
}
