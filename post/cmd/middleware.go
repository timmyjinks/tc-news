package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDContextKey contextKey = "userID"

// requireAuth validates a Bearer JWT issued by the auth service and injects
// the authenticated user's id into the request context. Handlers must read
// it via userIDFromContext -- no client-supplied header is trusted anymore.
func requireAuth(secret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(bearer, prefix) {
				http.Error(w, "Invalid user id", http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimPrefix(bearer, prefix)

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "Invalid user id", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid user id", http.StatusUnauthorized)
				return
			}
			userID, err := claims.GetSubject()
			if err != nil || userID == "" {
				http.Error(w, "Invalid user id", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next(w, r.WithContext(ctx))
		}
	}
}

func userIDFromContext(r *http.Request) string {
	if v, ok := r.Context().Value(userIDContextKey).(string); ok {
		return v
	}
	return ""
}
