package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/comment/store"
)

type application struct {
	store *store.PostgreStore
}

type CommentCreate struct {
	ParentId string `json:"parent_id"`
	Body     string `json:"body"`
}

type CommentUpdate struct {
	Body string `json:"body"`
}

func (app *application) Run(addr string) error {
	r := mux.NewRouter()

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	r.HandleFunc("/comments/{comment_id}", func(w http.ResponseWriter, r *http.Request) {
		commentId := mux.Vars(r)["comment_id"]

		if commentId == "" {
			http.Error(w, "Comment does not exist", http.StatusBadRequest)
			return
		}

		comment, err := app.store.GetById(commentId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(comment); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}).Methods("GET")

	r.HandleFunc("/posts/{post_id}/comments", func(w http.ResponseWriter, r *http.Request) {
		postId := mux.Vars(r)["post_id"]

		if postId == "" {
			http.Error(w, "Post does not exist", http.StatusBadRequest)
			return
		}

		comments, err := app.store.Get(postId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(comments); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}).Methods("GET")

	r.HandleFunc("/posts/{post_id}/comments", func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")
		postId := mux.Vars(r)["post_id"]

		if userId == "" {
			http.Error(w, "StatusUnauthorized", http.StatusUnauthorized)
			return
		}

		if postId == "" {
			http.Error(w, "Post does not exist", http.StatusBadRequest)
			return
		}

		var comment CommentCreate
		err := json.NewDecoder(r.Body).Decode(&comment)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := app.store.Create(store.CommentCreate{
			ParentId: comment.ParentId,
			PostId:   postId,
			UserId:   userId,
			Body:     comment.Body,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("POST")

	r.HandleFunc("/comments/{comment_id}", func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")
		commentId := mux.Vars(r)["comment_id"]

		if commentId == "" {
			http.Error(w, "Comment does not exist", http.StatusBadRequest)
			return
		}

		if userId == "" {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		}

		var comment CommentUpdate
		err := json.NewDecoder(r.Body).Decode(&comment)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := app.store.Update(store.CommentUpdate{
			Id:     commentId,
			UserId: userId,
			Body:   comment.Body,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("PUT")

	r.HandleFunc("/comments/{comment_id}", func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-User-ID")
		commentId := mux.Vars(r)["comment_id"]

		if commentId == "" {
			http.Error(w, "Comment does not exist", http.StatusBadRequest)
			return
		}

		if userId == "" {
			http.Error(w, "Invalid user id", http.StatusUnauthorized)
			return
		}

		if err := app.store.Delete(commentId, userId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}).Methods("DELETE")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
