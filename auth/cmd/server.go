package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/timmyjinks/auth/store"
)

var signingKey = []byte("sldfkjdskfjsdkfjdkjfskjfdkjfksjkdfjk")

type application struct {
	store *store.RedisStore
}

type Login struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (app *application) Run(addr string) error {
	r := mux.NewRouter()

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	r.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var login Login
		err := json.NewDecoder(r.Body).Decode(&login)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if login.Name == "" || login.Password == "" {
			http.Error(w, "name or password was empty", http.StatusBadRequest)
			return
		}

		data := store.Data{Name: login.Name, TTL: time.Hour * 24 * 7}
		refreshToken := generateRefreshToken()
		accessToken, err := generateAccessToken(store.Data{Name: login.Name, TTL: time.Minute * 15})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := app.store.Create(refreshToken, data); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Expires:  time.Now().Add(data.TTL),
		},
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AccessTokenResponse{AccessToken: accessToken})
	})

	r.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")
		refreshTokenCookie, err := r.Cookie("refresh_token")

		fmt.Println("b", bearer)

		const prefix = "Bearer "
		if !strings.HasPrefix(bearer, prefix) {
			http.Error(w, "not authorized user", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(bearer, prefix)
		fmt.Println("token", tokenString)

		exists, err := app.store.Exists(refreshTokenCookie.Value)
		if err != nil {
			if err := app.store.Delete(refreshTokenCookie.String()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !exists {
			http.Error(w, "unauthorized login again", http.StatusUnauthorized)
			return
		}

		if err := app.store.Delete(refreshTokenCookie.String()); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		data, err := verifyToken(tokenString)

		name, err := data.GetSubject()
		fmt.Println(name, "name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		refreshToken := generateRefreshToken()
		if err := app.store.Create(refreshToken, store.Data{Name: name, TTL: time.Minute * 15}); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		accessToken, err := generateAccessToken(store.Data{Name: name, TTL: time.Minute * 15})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Expires:  time.Now().Add(time.Hour * 24 * 7),
		},
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AccessTokenResponse{AccessToken: accessToken})
	})

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}

func generateRefreshToken() string {
	tokenBytes := make([]byte, 64)
	rand.Read(tokenBytes)
	return base64.RawURLEncoding.EncodeToString(tokenBytes)
}

func generateAccessToken(data store.Data) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "tysoncloud",
		Subject:   data.Name,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(data.TTL)),
	})
	return token.SignedString(signingKey)
}

func verifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
