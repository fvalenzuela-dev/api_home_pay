package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
)

type contextKey string

const AuthUserIDKey contextKey = "auth_user_id"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "no autorizado")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{Token: token})
		if err != nil {
			writeError(w, http.StatusUnauthorized, "no autorizado")
			return
		}

		ctx := context.WithValue(r.Context(), AuthUserIDKey, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetAuthUserID(r *http.Request) string {
	v, _ := r.Context().Value(AuthUserIDKey).(string)
	return v
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
