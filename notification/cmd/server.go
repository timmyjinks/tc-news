package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/notification/store"
)

type application struct {
	store *store.PostgreStore
}

type VoteInsert struct {
	Value int `json:"value"`
}

func (app *application) Run(addr string) error {
	r := mux.NewRouter()

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	r.HandleFunc("/user/notifications", func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")

		if userId == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		notifications, err := app.store.Get(userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(&notifications)
	}).Methods("GET")

	r.HandleFunc("/notifications/{notification_id}/read", func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")
		notificationId := mux.Vars(r)["notification_id"]

		if userId == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		err := app.store.Update(userId, notificationId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("PATCH")

	r.HandleFunc("/notifications/read-all", func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")

		if userId == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		err := app.store.UpdateAll(userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("PATCH")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
