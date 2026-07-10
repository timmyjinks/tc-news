package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SubscribeClient talks to the subscribe service over plain HTTP instead of
// gRPC. It's a drop-in replacement for grpcclient.SubscribeClient's
// GetSubscribers method, used to fan out notifications to a post's followers.
type SubscribeClient struct {
	baseURL string
	http    *http.Client
}

func NewSubscribeClient(baseURL string) *SubscribeClient {
	return &SubscribeClient{
		baseURL: baseURL,
		http:    &http.Client{},
	}
}

// GetSubscribers returns the user IDs following postId by calling the
// subscribe service's GET /posts/{post_id}/subscribers endpoint.
func (c *SubscribeClient) GetSubscribers(ctx context.Context, postId string) ([]string, error) {
	url := fmt.Sprintf("%s/posts/%s/subscribers", c.baseURL, postId)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscribe service returned status %d", resp.StatusCode)
	}

	var userIds []string
	if err := json.NewDecoder(resp.Body).Decode(&userIds); err != nil {
		return nil, err
	}
	return userIds, nil
}
