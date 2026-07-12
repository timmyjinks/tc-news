package main

import (
	"encoding/json"
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

// MarkNotificationRead godoc
// @Summary      Mark a notification as read
// @Description  Marks a single notification as read for the authenticated user
// @Tags         notifications
// @Param        notification_id  path      string  true  "Notification ID"
// @Param        X-User-ID        header    string  true  "ID of the authenticated user"
// @Success      200              "Notification marked as read"
// @Failure      401              {string}  string  "unauthorized"
// @Failure      500              {string}  string  "Internal server error"
// @Router       /notifications/{notification_id}/read [patch]
func (app *application) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	userId := userIDFromContext(r)
	notificationId := mux.Vars(r)["notification_id"]
	err := app.store.Update(notificationId, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// MarkAllNotificationsRead godoc
// @Summary      Mark all notifications as read
// @Description  Marks every notification as read for the authenticated user
// @Tags         notifications
// @Param        X-User-ID  header  string  true  "ID of the authenticated user"
// @Success      200        "All notifications marked as read"
// @Failure      401        {string}  string  "unauthorized"
// @Failure      500        {string}  string  "Internal server error"
// @Router       /notifications/read-all [patch]
func (app *application) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	userId := userIDFromContext(r)
	err := app.store.UpdateAll(userId)
	if err != nil {
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
	r.HandleFunc("/notifications/{notification_id}/read", auth(app.MarkNotificationRead)).Methods("PATCH")
	r.HandleFunc("/notifications/read-all", auth(app.MarkAllNotificationsRead)).Methods("PATCH")
	r.HandleFunc("/notifications/{notification_id}", auth(app.DeleteNotification)).Methods("DELETE")

	fmt.Printf("Listening on http://localhost%s\n", server.Addr)
	return server.ListenAndServe()
}
