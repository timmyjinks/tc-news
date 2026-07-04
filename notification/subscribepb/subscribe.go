package subscribepb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetSubscribersRequest struct {
	PostId string `json:"post_id"`
}

type GetSubscribersResponse struct {
	UserIds []string `json:"user_ids"`
}

// --- server ---

type SubscribeServiceServer interface {
	GetSubscribers(context.Context, *GetSubscribersRequest) (*GetSubscribersResponse, error)
}

type UnimplementedSubscribeServiceServer struct{}

func (UnimplementedSubscribeServiceServer) GetSubscribers(context.Context, *GetSubscribersRequest) (*GetSubscribersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetSubscribers not implemented")
}

func RegisterSubscribeServiceServer(s grpc.ServiceRegistrar, srv SubscribeServiceServer) {
	s.RegisterService(&subscribeServiceServiceDesc, srv)
}

func _SubscribeService_GetSubscribers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSubscribersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SubscribeServiceServer).GetSubscribers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/subscribe.SubscribeService/GetSubscribers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SubscribeServiceServer).GetSubscribers(ctx, req.(*GetSubscribersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var subscribeServiceServiceDesc = grpc.ServiceDesc{
	ServiceName: "subscribe.SubscribeService",
	HandlerType: (*SubscribeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSubscribers",
			Handler:    _SubscribeService_GetSubscribers_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "subscribe.proto",
}

// --- client ---

type SubscribeServiceClient interface {
	GetSubscribers(ctx context.Context, in *GetSubscribersRequest, opts ...grpc.CallOption) (*GetSubscribersResponse, error)
}

type subscribeServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSubscribeServiceClient(cc grpc.ClientConnInterface) SubscribeServiceClient {
	return &subscribeServiceClient{cc}
}

func (c *subscribeServiceClient) GetSubscribers(ctx context.Context, in *GetSubscribersRequest, opts ...grpc.CallOption) (*GetSubscribersResponse, error) {
	out := new(GetSubscribersResponse)
	if err := c.cc.Invoke(ctx, "/subscribe.SubscribeService/GetSubscribers", in, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}
