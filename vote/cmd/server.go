package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/vote/store"
)

// @title           Vote Service API
// @version         1.0
// @description     API for casting and removing votes on posts and comments.
// @host            localhost:8080
// @BasePath        /

type application struct {
	store *store.PostgreStore
}

type VoteInsert struct {
	Value int `json:"value"`
}

// ListUserVotes godoc
// @Summary      List a user's votes
// @Description  Retrieves all votes cast by the authenticated user
// @Tags         votes
// @Produce      json
// @Param        X-User-ID  header    string  true  "ID of the authenticated user"
// @Success      200        {array}   store.Vote
// @Failure      401        {string}  string  "Invalid user id"
// @Failure      500        {string}  string  "Internal server error"
// @Router       /user/votes [get]
func (app *application) ListUserVotes(w http.ResponseWriter, r *http.Request) {
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
}

// VotePost godoc
// @Summary      Vote on a post
// @Description  Casts or updates the authenticated user's vote on a post
// @Tags         votes
// @Accept       json
// @Param        post_id    path      string      true  "Post ID"
// @Param        X-User-ID  header    string      true  "ID of the authenticated user"
// @Param        vote       body      VoteInsert  true  "Vote payload"
// @Success      200        "Vote recorded"
// @Failure      400        {string}  string  "Invalid body"
// @Failure      401        {string}  string  "Post does not exist or invalid user id"
// @Failure      500        {string}  string  "Internal server error"
// @Router       /posts/{post_id}/votes [put]
func (app *application) VotePost(w http.ResponseWriter, r *http.Request) {
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
}

// VoteComment godoc
// @Summary      Vote on a comment
// @Description  Casts or updates the authenticated user's vote on a comment
// @Tags         votes
// @Accept       json
// @Param        comment_id  path      string      true  "Comment ID"
// @Param        X-User-ID   header    string      true  "ID of the authenticated user"
// @Param        vote        body      VoteInsert  true  "Vote payload"
// @Success      200         "Vote recorded"
// @Failure      400         {string}  string  "Invalid body"
// @Failure      401         {string}  string  "Invalid user id"
// @Failure      500         {string}  string  "Internal server error"
// @Router       /comments/{comment_id}/votes [put]
func (app *application) VoteComment(w http.ResponseWriter, r *http.Request) {
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
}

// DeletePostVote godoc
// @Summary      Remove a vote from a post
// @Description  Removes the authenticated user's vote on a post
// @Tags         votes
// @Param        post_id    path      string  true  "Post ID"
// @Param        X-User-ID  header    string  true  "ID of the authenticated user"
// @Success      200        "Vote removed"
// @Failure      401        {string}  string  "Invalid user id"
// @Failure      500        {string}  string  "Internal server error"
// @Router       /posts/{post_id}/votes [delete]
func (app *application) DeletePostVote(w http.ResponseWriter, r *http.Request) {
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
}

// DeleteCommentVote godoc
// @Summary      Remove a vote from a comment
// @Description  Removes the authenticated user's vote on a comment
// @Tags         votes
// @Param        comment_id  path      string  true  "Comment ID"
// @Param        X-User-ID   header    string  true  "ID of the authenticated user"
// @Success      200         "Vote removed"
// @Failure      401         {string}  string  "Invalid user id"
// @Failure      500         {string}  string  "Internal server error"
// @Router       /comments/{comment_id}/votes [delete]
func (app *application) DeleteCommentVote(w http.ResponseWriter, r *http.Request) {
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
}

func (app *application) Run(addr string) error {
	r := mux.NewRouter()
	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	r.HandleFunc("/user/votes", app.ListUserVotes).Methods("GET")
	r.HandleFunc("/posts/{post_id}/votes", app.VotePost).Methods("PUT")
	r.HandleFunc("/comments/{comment_id}/votes", app.VoteComment).Methods("PUT")
	r.HandleFunc("/posts/{post_id}/votes", app.DeletePostVote).Methods("DELETE")
	r.HandleFunc("/comments/{comment_id}/votes", app.DeleteCommentVote).Methods("DELETE")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
