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

// ErrorResponse is a generic error payload.
type ErrorResponse struct {
	Error string `json:"error"`
}

func (app *application) Run(addr string) error {
	r := mux.NewRouter()

	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	r.HandleFunc("/posts", app.ListPosts).Methods("GET")
	r.HandleFunc("/posts/{post_id}", app.GetPost).Methods("GET")
	r.HandleFunc("/posts", app.CreatePost).Methods("POST")
	r.HandleFunc("/posts/{post_id}", app.UpdatePost).Methods("PUT")
	r.HandleFunc("/posts/{post_id}", app.DeletePost).Methods("DELETE")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}

// ListPosts godoc
// @Summary      List all posts
// @Description  Returns every post
// @Tags         posts
// @Produce      json
// @Success      200  {array}   Post
// @Failure      500  {object}  ErrorResponse
// @Router       /posts [get]
func (app *application) ListPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := app.store.Get()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetPost godoc
// @Summary      Get a single post
// @Description  Returns one post by ID
// @Tags         posts
// @Produce      json
// @Param        post_id  path      string  true  "Post ID"
// @Success      200      {object}  Post
// @Failure      500      {object}  ErrorResponse
// @Router       /posts/{post_id} [get]
func (app *application) GetPost(w http.ResponseWriter, r *http.Request) {
	postId := mux.Vars(r)["post_id"]
	posts, err := app.store.GetById(postId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CreatePost godoc
// @Summary      Create a post
// @Description  Creates a new post for the requesting user
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        X-User-ID  header  string      true  "Requesting user's ID"
// @Param        post       body    PostCreate  true  "Post body"
// @Success      200        "post created"
// @Failure      400        {object}  ErrorResponse  "missing/malformed body"
// @Failure      401        {object}  ErrorResponse  "missing user id"
// @Failure      500        {object}  ErrorResponse
// @Router       /posts [post]
func (app *application) CreatePost(w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("X-User-ID")
	if userId == "" {
		http.Error(w, "Invalid user id", http.StatusUnauthorized)
		return
	}

	var post PostCreate
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := app.store.Create(store.PostCreate{
		UserId: userId,
		Body:   post.Body,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpdatePost godoc
// @Summary      Update a post
// @Description  Updates an existing post's body
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        X-User-ID  header  string      true  "Requesting user's ID"
// @Param        post_id    path    string      true  "Post ID"
// @Param        post       body    PostUpdate  true  "Updated post body"
// @Success      200        "post updated"
// @Failure      400        {object}  ErrorResponse  "missing post id or malformed body"
// @Failure      401        {object}  ErrorResponse  "missing user id"
// @Failure      500        {object}  ErrorResponse
// @Router       /posts/{post_id} [put]
func (app *application) UpdatePost(w http.ResponseWriter, r *http.Request) {
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

	var post PostUpdate
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
}

// DeletePost godoc
// @Summary      Delete a post
// @Description  Deletes a post owned by the requesting user
// @Tags         posts
// @Param        X-User-ID  header  string  true  "Requesting user's ID"
// @Param        post_id    path    string  true  "Post ID"
// @Success      200        "post deleted"
// @Failure      400        {object}  ErrorResponse  "missing post id"
// @Failure      401        {object}  ErrorResponse  "missing user id"
// @Failure      500        {object}  ErrorResponse
// @Router       /posts/{post_id} [delete]
func (app *application) DeletePost(w http.ResponseWriter, r *http.Request) {
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

	if err := app.store.Delete(postId, userId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
