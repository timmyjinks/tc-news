package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/post/store"
)

type application struct {
	store *store.PostgreStore
}

type Post struct {
	Id        string    `json:"id"`
	AuthorId  string    `json:"author_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

type PostCreate struct {
	Title string   `json:"title"`
	Body  string   `json:"body"`
	Tags  []string `json:"tags"`
}

type PostUpdate struct {
	Title string   `json:"title"`
	Body  string   `json:"body"`
	Tags  []string `json:"tags"`
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
// @Description  Returns posts, optionally paginated, filtered by tag, and sorted by creation date
// @Tags         posts
// @Produce      json
// @Param        limit   query     int     false  "Max number of posts to return"
// @Param        offset  query     int     false  "Number of posts to skip"
// @Param        tag     query     string  false  "Only return posts that have this tag"
// @Param        sort    query     string  false  "Sort order by creation date: newest (default) or oldest"
// @Success      200  {array}   Post
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /posts [get]
func (app *application) ListPosts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := 0
	if v := q.Get("limit"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	offset := 0
	if v := q.Get("offset"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 0 {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		offset = parsed
	}

	sort := q.Get("sort")
	if sort == "" {
		sort = "newest"
	}
	if sort != "newest" && sort != "oldest" {
		http.Error(w, "invalid sort: must be 'newest' or 'oldest'", http.StatusBadRequest)
		return
	}

	posts, err := app.store.Get(store.ListPostsParams{
		Limit:  limit,
		Offset: offset,
		Tag:    q.Get("tag"),
		Sort:   sort,
	})
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
// @Param        post       body    PostCreate  true  "Post title, body, and tags"
// @Success      200        "post created"
// @Failure      400        {object}  ErrorResponse  "missing/malformed body"
// @Failure      401        {object}  ErrorResponse  "missing user id"
// @Failure      500        {object}  ErrorResponse
// @Router       /posts [post]
func (app *application) CreatePost(w http.ResponseWriter, r *http.Request) {
	authorId := r.Header.Get("X-User-ID")
	if authorId == "" {
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
		AuthorId: authorId,
		Title:    post.Title,
		Body:     post.Body,
		Tags:     post.Tags,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpdatePost godoc
// @Summary      Update a post
// @Description  Updates an existing post's title, body, and tags
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        X-User-ID  header  string      true  "Requesting user's ID"
// @Param        post_id    path    string      true  "Post ID"
// @Param        post       body    PostUpdate  true  "Updated title, body, and tags"
// @Success      200        "post updated"
// @Failure      400        {object}  ErrorResponse  "missing post id or malformed body"
// @Failure      401        {object}  ErrorResponse  "missing user id"
// @Failure      500        {object}  ErrorResponse
// @Router       /posts/{post_id} [put]
func (app *application) UpdatePost(w http.ResponseWriter, r *http.Request) {
	authorId := r.Header.Get("X-User-ID")
	postId := mux.Vars(r)["post_id"]
	if authorId == "" {
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
		PostId:   postId,
		AuthorId: authorId,
		Title:    post.Title,
		Body:     post.Body,
		Tags:     post.Tags,
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
	authorId := r.Header.Get("X-User-ID")
	postId := mux.Vars(r)["post_id"]

	if authorId == "" {
		http.Error(w, "Invalid user id", http.StatusUnauthorized)
		return
	}
	if postId == "" {
		http.Error(w, "Post does not exist", http.StatusBadRequest)
		return
	}

	if err := app.store.Delete(postId, authorId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
