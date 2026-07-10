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

	r.HandleFunc("/users/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")

		if userId == "" {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		}

		subscriptions, err := app.store.Get(userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(subscriptions); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("GET")

	// GetSubscribers is the HTTP equivalent of the gRPC SubscribeService.GetSubscribers
	// call: it returns the plain list of user IDs following a post. Intended for
	// internal service-to-service use (e.g. by the notification service).
	r.HandleFunc("/posts/{post_id}/subscribers", func(w http.ResponseWriter, r *http.Request) {
		postId := mux.Vars(r)["post_id"]

		if postId == "" {
			http.Error(w, "Post does not exist", http.StatusBadRequest)
			return
		}

		subs, err := app.store.GetByPost(postId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userIds := make([]string, 0, len(subs))
		for _, sub := range subs {
			userIds = append(userIds, sub.UserId)
		}

		if err := json.NewEncoder(w).Encode(userIds); err != nil {
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

		err := app.store.Delete(postId, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("DELETE")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
