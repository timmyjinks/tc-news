package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/timmyjinks/follow/codec"
	"github.com/timmyjinks/follow/store"
	"github.com/timmyjinks/follow/subscribepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(codec.JSON{})
}

type server struct {
	subscribepb.UnimplementedSubscribeServiceServer
	store *store.PostgreStore
}

func (s *server) GetSubscribers(ctx context.Context, req *subscribepb.GetSubscribersRequest) (*subscribepb.GetSubscribersResponse, error) {
	subs, err := s.store.GetByPost(req.PostId)
	if err != nil {
		return nil, err
	}

	userIds := make([]string, 0, len(subs))
	for _, sub := range subs {
		userIds = append(userIds, sub.UserId)
	}

	return &subscribepb.GetSubscribersResponse{UserIds: userIds}, nil
}

// Serve starts the gRPC server and blocks until it stops or errors.
func Serve(addr string, s *store.PostgreStore) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	grpcServer := grpc.NewServer()
	subscribepb.RegisterSubscribeServiceServer(grpcServer, &server{store: s})

	log.Printf("gRPC listening on %s\n", addr)
	return grpcServer.Serve(lis)
}
