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

type CommentReplyCreatedPayload struct {
	CommentId       string `json:"comment_id"`
	PostId          string `json:"post_id"`
	UserId          string `json:"user_id"`
	ParentCommentId string `json:"parent_comment_id"`
	ParentAuthorId  string `json:"parent_author_id"`
	Body            string `json:"body"`
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

	case "comment_reply_created":
		var payload CommentReplyCreatedPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		// Don't notify someone for replying to their own comment.
		if payload.ParentAuthorId == payload.UserId {
			log.Println("[INFO] skipping self-reply notification for comment", payload.CommentId)
			return nil
		}

		if err := s.Create(payload.ParentAuthorId, "Someone replied to your comment"); err != nil {
			log.Println("[WARN] failed to create reply notification:", err)
			return err
		}
		log.Printf("[INFO] reply notification: comment=%s parent=%s notified=%s\n",
			payload.CommentId, payload.ParentCommentId, payload.ParentAuthorId)

	default:
		log.Println("[INFO] ignoring unknown message type:", msg.Type)
	}
	return nil
}
