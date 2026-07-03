package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/vote/store"
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

	r.HandleFunc("/user/votes", func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")

		if userId == "" {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		}

		votes := app.store.Get(userId)
		err := json.NewEncoder(w).Encode(votes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}).Methods("GET")

	r.HandleFunc("/posts/{post_id}/votes", func(w http.ResponseWriter, r *http.Request) {
		postId := mux.Vars(r)["post_id"]
		userId := r.Header.Get("X-User-ID")

		if postId == "" {
			http.Error(w, "Post does not exist", http.StatusUnauthorized)
			return
		}

		if userId == "" {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		}

		var vote VoteInsert
		err := json.NewDecoder(r.Body).Decode(&vote)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := app.store.InsertPost(store.VoteInsert{
			PostId: postId,
			UserId: userId,
			Value:  vote.Value,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("PUT")

	r.HandleFunc("/comments/{comment_id}/votes", func(w http.ResponseWriter, r *http.Request) {
		commentId := mux.Vars(r)["comment_id"]
		userId := r.Header.Get("X-User-ID")

		if userId == "" {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		}

		var vote VoteInsert
		err := json.NewDecoder(r.Body).Decode(&vote)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := app.store.InsertComment(store.VoteInsert{
			CommentId: commentId,
			UserId:    userId,
			Value:     vote.Value,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("PUT")

	r.HandleFunc("/posts/{post_id}/votes", func(w http.ResponseWriter, r *http.Request) {
		postId := mux.Vars(r)["post_id"]
		userId := r.Header.Get("X-User-ID")

		if userId == "" {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		}

		err := app.store.DeleteComment(postId, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("DELETE")

	r.HandleFunc("/comments/{comment_id}/votes", func(w http.ResponseWriter, r *http.Request) {
		commentId := mux.Vars(r)["comment_id"]
		userId := r.Header.Get("X-User-ID")

		if userId == "" {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		}

		err := app.store.DeleteComment(commentId, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("DELETE")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
