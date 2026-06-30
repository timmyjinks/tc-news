package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/user/store"
)

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

func (app *application) Run(addr string) error {
	r := mux.NewRouter()

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	r.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
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

	}).Methods("GET")

	r.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
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
	}).Methods("POST")

	r.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	}).Methods("PUT")

	r.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		userId := mux.Vars(r)["id"]
		if err := app.store.Delete(userId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("DELETE")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
