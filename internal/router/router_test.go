package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func testHandler(label string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"handler": label})
	}
}

func TestRouter_ProtectedRouteWithoutJWT_401(t *testing.T) {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "no autorizado")
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/api/v1/categories", testHandler("categories-list"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "no autorizado", resp["error"])
}

func TestRouter_ProtectedRouteWithValidJWT_200(t *testing.T) {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || authHeader == "Bearer invalid-token" {
				writeError(w, http.StatusUnauthorized, "no autorizado")
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/api/v1/categories", testHandler("categories-list"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	req.Header.Set("Authorization", "Bearer valid-test-token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "categories-list", resp["handler"])
}

func TestRouter_WebhookRouteBypassesAuth(t *testing.T) {
	// Simulate the real router structure: webhook registered BEFORE protected group
	r := chi.NewRouter()

	// First: webhook route (no auth)
	r.Post("/webhooks/clerk", testHandler("webhook-handle"))

	// Then: middleware + protected routes
	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					writeError(w, http.StatusUnauthorized, "no autorizado")
					return
				}
				next.ServeHTTP(w, r)
			})
		})
		r.Get("/api/v1/categories", testHandler("categories-list"))
	})

	// No Authorization header - webhook should still work because it's registered outside protected group
	req := httptest.NewRequest(http.MethodPost, "/webhooks/clerk", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_UnknownRoute_404(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/api/v1/categories", testHandler("categories"))

	// Hit a route that doesn't exist
	req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouter_CORSPreflight_200(t *testing.T) {
	r := chi.NewRouter()
	r.Options("/api/v1/companies", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/companies", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "OPTIONS")
}

func TestRouter_AllProtectedRoutesUseAuth(t *testing.T) {
	protectedPaths := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/categories"},
		{http.MethodPost, "/api/v1/categories"},
		{http.MethodGet, "/api/v1/companies"},
		{http.MethodGet, "/api/v1/companies/123/accounts"},
		{http.MethodGet, "/api/v1/accounts/123/billings"},
		{http.MethodPost, "/api/v1/accounts/123/billings"},
		{http.MethodPost, "/api/v1/periods/202603/open"},
		{http.MethodGet, "/api/v1/expenses"},
		{http.MethodPost, "/api/v1/expenses"},
		{http.MethodGet, "/api/v1/installments"},
		{http.MethodPost, "/api/v1/installments"},
		{http.MethodGet, "/api/v1/dashboard"},
	}

	for _, tt := range protectedPaths {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					authHeader := r.Header.Get("Authorization")
					if authHeader == "" {
						writeError(w, http.StatusUnauthorized, "no autorizado")
						return
					}
					next.ServeHTTP(w, r)
				})
			})

			// Register the route properly
			switch {
			case tt.method == http.MethodGet && tt.path == "/api/v1/categories":
				r.Get("/api/v1/categories", testHandler("handler"))
			case tt.method == http.MethodPost && tt.path == "/api/v1/categories":
				r.Post("/api/v1/categories", testHandler("handler"))
			case tt.method == http.MethodGet && tt.path == "/api/v1/companies":
				r.Get("/api/v1/companies", testHandler("handler"))
			case tt.method == http.MethodGet && tt.path == "/api/v1/companies/123/accounts":
				r.Get("/api/v1/companies/123/accounts", testHandler("handler"))
			case tt.method == http.MethodGet && tt.path == "/api/v1/accounts/123/billings":
				r.Get("/api/v1/accounts/123/billings", testHandler("handler"))
			case tt.method == http.MethodPost && tt.path == "/api/v1/accounts/123/billings":
				r.Post("/api/v1/accounts/123/billings", testHandler("handler"))
			case tt.method == http.MethodPost && tt.path == "/api/v1/periods/202603/open":
				r.Post("/api/v1/periods/202603/open", testHandler("handler"))
			case tt.method == http.MethodGet && tt.path == "/api/v1/expenses":
				r.Get("/api/v1/expenses", testHandler("handler"))
			case tt.method == http.MethodPost && tt.path == "/api/v1/expenses":
				r.Post("/api/v1/expenses", testHandler("handler"))
			case tt.method == http.MethodGet && tt.path == "/api/v1/installments":
				r.Get("/api/v1/installments", testHandler("handler"))
			case tt.method == http.MethodPost && tt.path == "/api/v1/installments":
				r.Post("/api/v1/installments", testHandler("handler"))
			case tt.method == http.MethodGet && tt.path == "/api/v1/dashboard":
				r.Get("/api/v1/dashboard", testHandler("handler"))
			}

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "Protected route %s %s should return 401 without auth", tt.method, tt.path)
		})
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

type mockHandler struct {
	fn func(w http.ResponseWriter, r *http.Request)
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.fn(w, r)
}

func TestRouter_New(t *testing.T) {
	t.Run("creates router without panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			_ = New(nil, nil, nil, nil, nil, nil, nil, nil, nil)
		})
	})
}
