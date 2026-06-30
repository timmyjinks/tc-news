package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/timmyjinks/auth/store"
)

type application struct {
	store  *store.RedisStore
	pstore *store.PostgresStore
}

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (app *application) Run(addr string) error {
	r := mux.NewRouter()

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	r.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := app.pstore.GetByNameAndPassword(user.Name, user.Password); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

	}).Methods("POST")

	r.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {

	}).Methods("POST")

	r.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		refreshToken := cookie.Value

		if refreshToken == "" {
			http.Error(w, "no refresh token found", http.StatusUnauthorized)
			return
		}
		session, err := s.store.GetRefreshToken(refreshToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if time.Now().After(session.ExpiresAt) {
			s.store.DeleteRefreshToken(session.RefreshToken)
			http.Error(w, "refreshToken expired please login", http.StatusUnauthorized)
			return
		}

		if err := s.store.DeleteRefreshToken(refreshToken); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newRefreshToken := GenerateRefreshToken()
		if err := s.store.CreateRefreshToken(session.UserId, newRefreshToken); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newToken, err := CreateJWTToken(session.UserId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    newToken,
			HttpOnly: true,
			Secure:   true,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			HttpOnly: true,
			Secure:   true,
		})
	}).Methods("POST")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
