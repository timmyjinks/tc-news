package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/timmyjinks/notification/store"
)

type application struct {
	store     *store.PostgreStore
	jwtSecret string
}

type VoteInsert struct {
	Value int `json:"value"`
}

type StatusUpdate struct {
	Status string `json:"status"`
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
func (app *application) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userId := userIDFromContext(r)
	notifications, err := app.store.Get(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(&notifications)
}

// UpdateNotificationStatus godoc
// @Summary      Transition a notification's status
// @Description  Validates the transition against the notification state machine and applies it
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        notification_id  path      string        true  "Notification ID"
// @Param        X-User-ID        header    string        true  "ID of the authenticated user"
// @Param        status           body      StatusUpdate  true  "Target status"
// @Success      200              {object}  store.Notification
// @Failure      400              {string}  string  "Invalid status or illegal transition"
// @Failure      401              {string}  string  "unauthorized"
// @Failure      500              {string}  string  "Internal server error"
// @Router       /notifications/{notification_id}/status [patch]
func (app *application) UpdateNotificationStatus(w http.ResponseWriter, r *http.Request) {
	userId := userIDFromContext(r)
	notificationId := mux.Vars(r)["notification_id"]

	var body StatusUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if !isValidStatus(body.Status) {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	current, err := app.store.GetStatus(notificationId, userId)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, "Notification does not exist", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !canTransition(current, body.Status) {
		http.Error(w, fmt.Sprintf("illegal transition: %s -> %s", current, body.Status), http.StatusBadRequest)
		return
	}

	updated, err := app.store.UpdateStatus(notificationId, userId, body.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(updated)
}

// MarkAllNotificationsRead godoc
// @Summary      Mark all notifications as read
// @Description  Transitions every DELIVERED notification for the authenticated user to READ
// @Tags         notifications
// @Param        X-User-ID  header  string  true  "ID of the authenticated user"
// @Success      200        "All eligible notifications marked as read"
// @Failure      401        {string}  string  "unauthorized"
// @Failure      500        {string}  string  "Internal server error"
// @Router       /notifications/read-all [patch]
func (app *application) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	userId := userIDFromContext(r)
	if err := app.store.MarkAllRead(userId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteNotification godoc
// @Summary      Delete a notification
// @Description  Permanently deletes a notification owned by the authenticated user
// @Tags         notifications
// @Param        notification_id  path      string  true  "Notification ID"
// @Param        X-User-ID        header    string  true  "ID of the authenticated user"
// @Success      200              "Notification deleted"
// @Failure      400              {string}  string  "Notification does not exist"
// @Failure      401              {string}  string  "unauthorized"
// @Failure      500              {string}  string  "Internal server error"
// @Router       /notifications/{notification_id} [delete]
func (app *application) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	userId := userIDFromContext(r)
	notificationId := mux.Vars(r)["notification_id"]
	if notificationId == "" {
		http.Error(w, "Notification does not exist", http.StatusBadRequest)
		return
	}
	if err := app.store.Delete(notificationId, userId); err != nil {
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

	auth := requireAuth(app.jwtSecret)
	r.HandleFunc("/user/notifications", auth(app.ListNotifications)).Methods("GET")
	r.HandleFunc("/notifications/{notification_id}/status", auth(app.UpdateNotificationStatus)).Methods("PATCH")
	r.HandleFunc("/notifications/read-all", auth(app.MarkAllNotificationsRead)).Methods("PATCH")
	r.HandleFunc("/notifications/{notification_id}", auth(app.DeleteNotification)).Methods("DELETE")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
