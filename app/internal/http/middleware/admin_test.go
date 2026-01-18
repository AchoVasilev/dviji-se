package middleware

import (
	"net/http"
	"net/http/httptest"
	"server/internal/domain/user"
	"server/util/ctxutils"
	"server/util/securityutil"
	"testing"

	"github.com/google/uuid"
)

func TestRequireAuth_Authenticated(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := RequireAuth(next)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	user := &securityutil.LoggedInUser{
		Id:       uuid.New().String(),
		Username: "test@example.com",
	}
	ctx := ctxutils.WithLoggedUser(req.Context(), user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !nextCalled {
		t.Error("RequireAuth should call next handler for authenticated user")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRequireAuth_NotAuthenticated(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := RequireAuth(next)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if nextCalled {
		t.Error("RequireAuth should not call next handler for unauthenticated user")
	}

	if w.Code != http.StatusSeeOther {
		t.Errorf("Status = %d, want %d (redirect)", w.Code, http.StatusSeeOther)
	}

	location := w.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Redirect location = %q, want /login", location)
	}
}

func TestRequireAuth_NilUser(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := RequireAuth(next)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	// Set nil user explicitly
	ctx := ctxutils.WithLoggedUser(req.Context(), nil)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if nextCalled {
		t.Error("RequireAuth should not call next handler for nil user")
	}

	if w.Code != http.StatusSeeOther {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusSeeOther)
	}
}

func TestRequireAdmin_AdminUser(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := RequireAdmin(next)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	adminUser := &securityutil.LoggedInUser{
		Id:       uuid.New().String(),
		Username: "admin@example.com",
		Roles: []user.Role{
			{Id: uuid.New(), Name: "ADMIN"},
			{Id: uuid.New(), Name: "USER"},
		},
	}
	ctx := ctxutils.WithLoggedUser(req.Context(), adminUser)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !nextCalled {
		t.Error("RequireAdmin should call next handler for admin user")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRequireAdmin_NonAdminUser(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := RequireAdmin(next)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	regularUser := &securityutil.LoggedInUser{
		Id:       uuid.New().String(),
		Username: "user@example.com",
		Roles: []user.Role{
			{Id: uuid.New(), Name: "USER"},
		},
	}
	ctx := ctxutils.WithLoggedUser(req.Context(), regularUser)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if nextCalled {
		t.Error("RequireAdmin should not call next handler for non-admin user")
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestRequireAdmin_NoRoles(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := RequireAdmin(next)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	userNoRoles := &securityutil.LoggedInUser{
		Id:       uuid.New().String(),
		Username: "noroles@example.com",
		Roles:    []user.Role{},
	}
	ctx := ctxutils.WithLoggedUser(req.Context(), userNoRoles)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if nextCalled {
		t.Error("RequireAdmin should not call next handler for user with no roles")
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestRequireAdmin_NotAuthenticated(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := RequireAdmin(next)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if nextCalled {
		t.Error("RequireAdmin should not call next handler for unauthenticated user")
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestRequireAdmin_NilUser(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := RequireAdmin(next)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := ctxutils.WithLoggedUser(req.Context(), nil)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if nextCalled {
		t.Error("RequireAdmin should not call next handler for nil user")
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestRequireAdmin_AdminRoleCaseSensitive(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := RequireAdmin(next)

	tests := []struct {
		name     string
		roleName string
		expected bool
	}{
		{"uppercase ADMIN", "ADMIN", true},
		{"lowercase admin", "admin", false},
		{"mixed case Admin", "Admin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled = false
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			user := &securityutil.LoggedInUser{
				Id:       uuid.New().String(),
				Username: "test@example.com",
				Roles: []user.Role{
					{Id: uuid.New(), Name: tt.roleName},
				},
			}
			ctx := ctxutils.WithLoggedUser(req.Context(), user)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if nextCalled != tt.expected {
				t.Errorf("RequireAdmin for role %q: nextCalled = %v, want %v", tt.roleName, nextCalled, tt.expected)
			}
		})
	}
}
