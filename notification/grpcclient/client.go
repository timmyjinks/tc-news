package grpcclient

import (
	"context"

	"github.com/timmyjinks/notification/codec"
	"github.com/timmyjinks/notification/subscribepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(codec.JSON{})
}

type SubscribeClient struct {
	client subscribepb.SubscribeServiceClient
}

func NewSubscribeClient(addr string) (*SubscribeClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(codec.Name)),
	)
	if err != nil {
		return nil, err
	}
	return &SubscribeClient{client: subscribepb.NewSubscribeServiceClient(conn)}, nil
}

// GetSubscribers returns the user IDs following postId.
func (c *SubscribeClient) GetSubscribers(ctx context.Context, postId string) ([]string, error) {
	resp, err := c.client.GetSubscribers(ctx, &subscribepb.GetSubscribersRequest{PostId: postId})
	if err != nil {
		return nil, err
	}
	return resp.UserIds, nil
}
