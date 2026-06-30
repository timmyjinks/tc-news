package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/auth/store"
)

type application struct {
	store *store.RedisStore
}

type Login struct {
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

	}).Methods("POST")

	r.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {

	}).Methods("POST")

	r.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {

	}).Methods("POST")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
