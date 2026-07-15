package main

import (
	"encoding/json"
	"fmt"
	"log"

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

func handleMessage(s *store.PostgreStore, msg kafka.Message) error {
	switch msg.Type {
	case "comment_created":
		var payload CommentCreatedPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}
		if err := s.Create(msg.Type, fmt.Sprintf("%s created comment on post %s", payload.UserId, payload.PostId)); err != nil {
			log.Println("[WARN] failed to create comment created event:", err)
			return err
		}
	case "comment_reply_created":
		var payload CommentReplyCreatedPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}

		if err := s.Create(msg.Type, fmt.Sprintf("%s replied to %s", payload.UserId, payload.ParentAuthorId)); err != nil {
			log.Println("[WARN] failed to create reply event:", err)
			return err
		}

	default:
		log.Println("[INFO] ignoring unknown message type:", msg.Type)
	}
	return nil
}
