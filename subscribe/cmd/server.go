package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/follow/store"
)

type application struct {
	store *store.PostgreStore
}

func (app *application) Run(addr string) error {
	r := mux.NewRouter()

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	r.HandleFunc("/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("healthy"))
	}).Methods("GET")

	r.HandleFunc("/users/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")

		if userId == "" {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		}

		subscribers, err := app.store.Get(userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(subscribers); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}).Methods("GET")

	r.HandleFunc("/posts/{post_id}/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")
		postId := mux.Vars(r)["post_id"]

		if userId == "" {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		}

		if postId == "" {
			http.Error(w, "Post does not exist", http.StatusBadRequest)
			return
		}

		err := app.store.Create(store.SubscriberCreate{
			PostId: postId,
			UserId: userId,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("POST")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
