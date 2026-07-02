package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/post/store"
)

type application struct {
	store *store.PostgreStore
}

type Post struct {
	Id        string    `json:"id"`
	UserId    string    `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type PostCreate struct {
	Body string `json:"body"`
}

type PostUpdate struct {
	Body string `json:"body"`
}

func (app *application) Run(addr string) error {
	r := mux.NewRouter()

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	r.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		posts, err := app.store.Get()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(posts); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}).Methods("GET")

	r.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")

		var post PostCreate
		err := json.NewDecoder(r.Body).Decode(&post)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if userId == "" {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		}

		if err := app.store.Create(store.PostCreate{
			UserId: userId,
			Body:   post.Body,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("POST")

	r.HandleFunc("/posts/{id}", func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")
		postId := mux.Vars(r)["id"]

		var post PostCreate
		err := json.NewDecoder(r.Body).Decode(&post)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if userId == "" {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		}

		if err := app.store.Update(store.PostUpdate{
			PostId: postId,
			UserId: userId,
			Body:   post.Body,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("PUT")

	r.HandleFunc("/posts/{id}", func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")
		postId := mux.Vars(r)["id"]

		var post PostCreate
		err := json.NewDecoder(r.Body).Decode(&post)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if userId == "" {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		}

		if err := app.store.Delete(postId, userId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("DELETE")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
