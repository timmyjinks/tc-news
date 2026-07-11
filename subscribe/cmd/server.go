package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/follow/store"
)

type application struct {
	store     *store.PostgreStore
	jwtSecret string
}

func (app *application) Run(addr string) error {
	r := mux.NewRouter()

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	auth := requireAuth(app.jwtSecret)

	r.HandleFunc("/users/subscriptions", auth(func(w http.ResponseWriter, r *http.Request) {
		userId := userIDFromContext(r)

		subscriptions, err := app.store.Get(userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(subscriptions); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})).Methods("GET")

	// Internal service-to-service call (used by notification) -- not user
	// facing, so it stays outside the auth middleware.
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

	r.HandleFunc("/posts/{post_id}/subscriptions", auth(func(w http.ResponseWriter, r *http.Request) {
		userId := userIDFromContext(r)
		postId := mux.Vars(r)["post_id"]

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
	})).Methods("POST")

	r.HandleFunc("/posts/{post_id}/subscriptions", auth(func(w http.ResponseWriter, r *http.Request) {
		userId := userIDFromContext(r)
		postId := mux.Vars(r)["post_id"]

		if postId == "" {
			http.Error(w, "Post does not exist", http.StatusBadRequest)
			return
		}

		err := app.store.Delete(postId, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})).Methods("DELETE")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
