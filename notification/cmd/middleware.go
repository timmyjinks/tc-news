package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func requireAuth(secret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(bearer, prefix) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
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
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			userID, err := claims.GetSubject()
			if err != nil || userID == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
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
