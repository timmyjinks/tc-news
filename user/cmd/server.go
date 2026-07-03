package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/user/store"
)

// @title           User Service API
// @version         1.0
// @description     API for creating, reading, updating, and deleting users.
// @host            localhost:8080
// @BasePath        /

type application struct {
	store *store.PostgreStore
}

type User struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type UserCreate struct {
	Name string `json:"name"`
}

type UserUpdate struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type UserDelete struct {
	Id string `json:"id"`
}

// GetUser godoc
// @Summary      Get a user by ID
// @Description  Retrieves a single user by their ID
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  User
// @Failure      401  {string}  string  "Invalid user id"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /users/{id} [get]
func (app *application) GetUser(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["id"]
	if userId == "" {
		http.Error(w, "Invalid user id", http.StatusUnauthorized)
		return
	}
	user, err := app.store.GetById(userId)
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
// @Summary      Create a user
// @Description  Creates a new user
// @Tags         users
// @Accept       json
// @Param        user  body  UserCreate  true  "User payload"
// @Success      201   "User created"
// @Failure      400   {string}  string  "Invalid body"
// @Failure      500   {string}  string  "Internal server error"
// @Router       /users [post]
func (app *application) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user UserCreate
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := app.store.Create(store.UserCreate{
		Name: user.Name,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(201)
}

// UpdateUser godoc
// @Summary      Update a user
// @Description  Updates an existing user's name
// @Tags         users
// @Accept       json
// @Param        id    path  string      true  "User ID"
// @Param        user  body  UserCreate  true  "Updated user payload"
// @Success      200   "User updated"
// @Failure      400   {string}  string  "Invalid body"
// @Failure      500   {string}  string  "Internal server error"
// @Router       /users/{id} [put]
func (app *application) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["id"]
	var user UserCreate
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := app.store.Update(store.UserUpdate{
		Id:   userId,
		Name: user.Name,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Deletes a user by ID
// @Tags         users
// @Param        id   path  string  true  "User ID"
// @Success      200  "User deleted"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /users/{id} [delete]
func (app *application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["id"]
	if err := app.store.Delete(userId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (app *application) Run(addr string) error {
	r := mux.NewRouter()
	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	r.HandleFunc("/users/{id}", app.GetUser).Methods("GET")
	r.HandleFunc("/users", app.CreateUser).Methods("POST")
	r.HandleFunc("/users/{id}", app.UpdateUser).Methods("PUT")
	r.HandleFunc("/users/{id}", app.DeleteUser).Methods("DELETE")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
