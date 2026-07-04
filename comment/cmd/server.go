package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/comment/kafka"
	"github.com/timmyjinks/comment/store"
)

type application struct {
	store    *store.PostgreStore
	producer *kafka.KafkaService
}

type CommentCreate struct {
	ParentId string `json:"parent_id"`
	Body     string `json:"body"`
}

type CommentUpdate struct {
	Body string `json:"body"`
}

// GetComment godoc
// @Summary      Get a comment by ID
// @Description  Retrieves a single comment by its ID
// @Tags         comments
// @Produce      json
// @Param        comment_id  path      string  true  "Comment ID"
// @Success      200         {object}  store.Comment
// @Failure      400         {string}  string  "Comment does not exist"
// @Failure      500         {string}  string  "Internal server error"
// @Router       /comments/{comment_id} [get]
func (app *application) GetComment(w http.ResponseWriter, r *http.Request) {
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
}

// ListComments godoc
// @Summary      List comments for a post
// @Description  Retrieves all comments belonging to a post
// @Tags         comments
// @Produce      json
// @Param        post_id  path      string  true  "Post ID"
// @Success      200      {array}   store.Comment
// @Failure      400      {string}  string  "Post does not exist"
// @Failure      500      {string}  string  "Internal server error"
// @Router       /posts/{post_id}/comments [get]
func (app *application) ListComments(w http.ResponseWriter, r *http.Request) {
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
}

// CreateComment godoc
// @Summary      Create a comment
// @Description  Creates a new comment (optionally a reply, via parent_id) on a post
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        post_id     path      string          true  "Post ID"
// @Param        X-User-ID   header    string          true  "ID of the authenticated user"
// @Param        comment     body      CommentCreate   true  "Comment payload"
// @Success      201         "Comment created"
// @Failure      400         {string}  string  "Post does not exist or invalid body"
// @Failure      401         {string}  string  "Unauthorized"
// @Failure      500         {string}  string  "Internal server error"
// @Router       /posts/{post_id}/comments [post]
func (app *application) CreateComment(w http.ResponseWriter, r *http.Request) {
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
	commentId, err := app.store.Create(store.CommentCreate{
		ParentId: comment.ParentId,
		PostId:   postId,
		UserId:   userId,
		Body:     comment.Body,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(struct {
		CommentId string `json:"comment_id"`
		PostId    string `json:"post_id"`
		UserId    string `json:"user_id"`
		Body      string `json:"body"`
	}{CommentId: commentId, PostId: postId, UserId: userId, Body: comment.Body})
	if err != nil {
		log.Println("[WARN] failed to marshal comment_created payload:", err)
		return
	}

	if err := app.producer.Producer.Send(r.Context(), "notifications", kafka.Message{
		Type:    "comment_created",
		Payload: payload,
	}); err != nil {
		log.Println("[WARN] failed to publish comment_created event:", err)
	}
}

// UpdateComment godoc
// @Summary      Update a comment
// @Description  Updates the body of an existing comment
// @Tags         comments
// @Accept       json
// @Param        comment_id  path      string          true  "Comment ID"
// @Param        X-User-ID   header    string          true  "ID of the authenticated user"
// @Param        comment     body      CommentUpdate   true  "Updated comment payload"
// @Success      200         "Comment updated"
// @Failure      400         {string}  string  "Comment does not exist or invalid body"
// @Failure      401         {string}  string  "Invalid user id"
// @Failure      500         {string}  string  "Internal server error"
// @Router       /comments/{comment_id} [put]
func (app *application) UpdateComment(w http.ResponseWriter, r *http.Request) {
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
}

// DeleteComment godoc
// @Summary      Delete a comment
// @Description  Deletes a comment owned by the authenticated user
// @Tags         comments
// @Param        comment_id  path      string  true  "Comment ID"
// @Param        X-User-ID   header    string  true  "ID of the authenticated user"
// @Success      200         "Comment deleted"
// @Failure      400         {string}  string  "Comment does not exist"
// @Failure      401         {string}  string  "Invalid user id"
// @Failure      500         {string}  string  "Internal server error"
// @Router       /comments/{comment_id} [delete]
func (app *application) DeleteComment(w http.ResponseWriter, r *http.Request) {
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
}

func (app *application) Run(addr string) error {
	r := mux.NewRouter()
	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	r.HandleFunc("/comments/{comment_id}", app.GetComment).Methods("GET")
	r.HandleFunc("/posts/{post_id}/comments", app.ListComments).Methods("GET")
	r.HandleFunc("/posts/{post_id}/comments", app.CreateComment).Methods("POST")
	r.HandleFunc("/comments/{comment_id}", app.UpdateComment).Methods("PUT")
	r.HandleFunc("/comments/{comment_id}", app.DeleteComment).Methods("DELETE")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
