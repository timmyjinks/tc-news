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
	"golang.org/x/crypto/bcrypt"
)

type application struct {
	store     *store.RedisStore
	userStore *store.PostgresStore
	jwtSecret string
}

type Login struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type User struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type UserCreate struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (app *application) Run(addr string) error {
	r := mux.NewRouter()
	server := http.Server{Addr: addr, Handler: r}

	r.HandleFunc("/auth/login", app.Login).Methods(http.MethodPost)
	r.HandleFunc("/auth/refresh", app.Refresh).Methods(http.MethodPost)

	r.HandleFunc("/users/{id}", app.GetUser).Methods(http.MethodGet)
	r.HandleFunc("/users", app.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/users/{id}", app.UpdateUser).Methods(http.MethodPut)
	r.HandleFunc("/users/{id}", app.DeleteUser).Methods(http.MethodDelete)

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}

// Login godoc
// @Summary      Log in with a username and password
// @Description  Validates credentials against stored user records, issues a short-lived access token, and sets a refresh token cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        login  body      Login  true  "Login credentials"
// @Success      200    {object}  AccessTokenResponse
// @Failure      400    {object}  ErrorResponse  "missing or malformed body"
// @Failure      401    {object}  ErrorResponse  "invalid credentials"
// @Failure      500    {object}  ErrorResponse  "internal error"
// @Router       /auth/login [post]
func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	var login Login
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if login.Name == "" || login.Password == "" {
		http.Error(w, "name or password was empty", http.StatusBadRequest)
		return
	}

	user, err := app.userStore.GetByName(login.Name)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(login.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	data := store.Data{Id: user.Id, Name: user.Name, TTL: time.Hour * 24 * 7}
	refreshToken := generateRefreshToken()
	accessToken, err := app.generateAccessToken(store.Data{Id: user.Id, Name: user.Name, TTL: time.Minute * 15})
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
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AccessTokenResponse{AccessToken: accessToken})
}

// Refresh godoc
// @Summary      Refresh an access token
// @Description  Exchanges a valid refresh token cookie + bearer access token for a new access/refresh token pair
// @Tags         auth
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer access token"
// @Success      200            {object}  AccessTokenResponse
// @Failure      401            {object}  ErrorResponse  "not authorized / session expired"
// @Failure      500            {object}  ErrorResponse  "internal error"
// @Router       /auth/refresh [post]
func (app *application) Refresh(w http.ResponseWriter, r *http.Request) {
	bearer := r.Header.Get("Authorization")
	refreshTokenCookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "not authorized user", http.StatusUnauthorized)
		return
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(bearer, prefix) {
		http.Error(w, "not authorized user", http.StatusUnauthorized)
		return
	}
	tokenString := strings.TrimPrefix(bearer, prefix)

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

	claims, err := app.verifyToken(tokenString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userId, err := claims.GetSubject()
	if err != nil || userId == "" {
		http.Error(w, "invalid token subject", http.StatusUnauthorized)
		return
	}

	// Re-fetch the user so a deleted/renamed account can't keep refreshing,
	// and so the reissued token carries an up-to-date Name.
	user, err := app.userStore.GetById(userId)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	refreshToken := generateRefreshToken()
	if err := app.store.Create(refreshToken, store.Data{Id: user.Id, Name: user.Name, TTL: time.Hour * 24 * 7}); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	accessToken, err := app.generateAccessToken(store.Data{Id: user.Id, Name: user.Name, TTL: time.Minute * 15})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AccessTokenResponse{AccessToken: accessToken})
}

// GetUser godoc
// @Router       /users/{id} [get]
func (app *application) GetUser(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["id"]
	if userId == "" {
		http.Error(w, "Invalid user id", http.StatusUnauthorized)
		return
	}
	user, err := app.userStore.GetById(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(User{
		Id:        user.Id,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CreateUser godoc
// @Description  Creates a new user with a bcrypt-hashed password
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        login  body      Login  true  "Login credentials"
// @Success      200    {object}  AccessTokenResponse
// @Failure      400    {object}  ErrorResponse  "missing or malformed body"
// @Failure      401    {object}  ErrorResponse  "invalid credentials"
// @Failure      500    {object}  ErrorResponse  "internal error"
// @Router       /users [post]
func (app *application) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user UserCreate
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Name == "" || user.Password == "" {
		http.Error(w, "name or password was empty", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := app.userStore.Create(store.UserCreate{
		Name:         user.Name,
		PasswordHash: string(hash),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(201)
}

// UpdateUser godoc
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        login  body      Login  true  "Login credentials"
// @Success      200    {object}  AccessTokenResponse
// @Failure      400    {object}  ErrorResponse  "missing or malformed body"
// @Failure      401    {object}  ErrorResponse  "invalid credentials"
// @Failure      500    {object}  ErrorResponse  "internal error"
// @Router       /users/{id} [put]
func (app *application) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["id"]
	var user UserCreate
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := app.userStore.Update(store.UserUpdate{
		Id:   userId,
		Name: user.Name,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteUser godoc
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        login  body      Login  true  "Login credentials"
// @Success      200    {object}  AccessTokenResponse
// @Failure      400    {object}  ErrorResponse  "missing or malformed body"
// @Failure      401    {object}  ErrorResponse  "invalid credentials"
// @Failure      500    {object}  ErrorResponse  "internal error"
// @Router       /users/{id} [delete]
func (app *application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["id"]
	if err := app.userStore.Delete(userId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func generateRefreshToken() string {
	tokenBytes := make([]byte, 64)
	rand.Read(tokenBytes)
	return base64.RawURLEncoding.EncodeToString(tokenBytes)
}

func (app *application) generateAccessToken(data store.Data) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "tysoncloud",
		Subject:   data.Id,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(data.TTL)),
	})
	return token.SignedString([]byte(app.jwtSecret))
}

func (app *application) verifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(app.jwtSecret), nil
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
