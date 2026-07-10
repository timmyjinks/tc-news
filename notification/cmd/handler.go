package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/timmyjinks/notification/httpclient"
	"github.com/timmyjinks/notification/kafka"
	"github.com/timmyjinks/notification/store"
)

type CommentCreatedPayload struct {
	CommentId string `json:"comment_id"`
	PostId    string `json:"post_id"`
	UserId    string `json:"user_id"`
	Body      string `json:"body"`
}

func handleMessage(s *store.PostgreStore, subscribeClient *httpclient.SubscribeClient, msg kafka.Message) error {
	switch msg.Type {
	case "comment_created":
		var payload CommentCreatedPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		userIds, err := subscribeClient.GetSubscribers(context.Background(), payload.PostId)
		if err != nil {
			return err
		}

		for _, userId := range userIds {
			if userId == payload.UserId {
				continue
			}
			if err := s.Create(userId, "New comment on a post you follow"); err != nil {
				log.Println("[WARN] failed to create notification:", err)
			}
		}
	default:
		log.Println("[INFO] ignoring unknown message type:", msg.Type)
	}
	return nil
}
