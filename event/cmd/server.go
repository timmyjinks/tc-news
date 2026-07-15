package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/notification/store"
)

type application struct {
	store *store.PostgreStore
}

// ListNotifications godoc
// @Summary      List a user's notifications
// @Description  Retrieves all notifications for the authenticated user
// @Tags         notifications
// @Produce      json
// @Param        X-User-ID  header    string  true  "ID of the authenticated user"
// @Success      200        {array}   store.Notification
// @Failure      401        {string}  string  "unauthorized"
// @Failure      500        {string}  string  "Internal server error"
// @Router       /user/notifications [get]
func (app *application) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := app.store.Get()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(&events)
}

func (app *application) Run(addr string) error {
	r := mux.NewRouter()
	server := http.Server{
		Addr:    addr,
		Handler: r,
	}

	r.HandleFunc("/events", app.ListEvents).Methods("GET")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
