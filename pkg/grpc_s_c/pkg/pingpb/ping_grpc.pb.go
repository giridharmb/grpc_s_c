// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: proto/ping.proto

package pingpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	PingService_SendPing_FullMethodName          = "/ping.PingService/SendPing"
	PingService_ListenForMessages_FullMethodName = "/ping.PingService/ListenForMessages"
)

// PingServiceClient is the client API for PingService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// PingService defines two RPC methods:
//   - SendPing: Unary RPC used by a client (virtual machine) to send its status along with a JSON payload.
//   - ListenForMessages: Server‑streaming RPC that pushes messages (with a JSON payload) to the client.
type PingServiceClient interface {
	SendPing(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	ListenForMessages(ctx context.Context, in *ListenRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[Message], error)
}

type pingServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPingServiceClient(cc grpc.ClientConnInterface) PingServiceClient {
	return &pingServiceClient{cc}
}

func (c *pingServiceClient) SendPing(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, PingService_SendPing_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pingServiceClient) ListenForMessages(ctx context.Context, in *ListenRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[Message], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &PingService_ServiceDesc.Streams[0], PingService_ListenForMessages_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[ListenRequest, Message]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type PingService_ListenForMessagesClient = grpc.ServerStreamingClient[Message]

// PingServiceServer is the server API for PingService service.
// All implementations must embed UnimplementedPingServiceServer
// for forward compatibility.
//
// PingService defines two RPC methods:
//   - SendPing: Unary RPC used by a client (virtual machine) to send its status along with a JSON payload.
//   - ListenForMessages: Server‑streaming RPC that pushes messages (with a JSON payload) to the client.
type PingServiceServer interface {
	SendPing(context.Context, *PingRequest) (*PingResponse, error)
	ListenForMessages(*ListenRequest, grpc.ServerStreamingServer[Message]) error
	mustEmbedUnimplementedPingServiceServer()
}

// UnimplementedPingServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedPingServiceServer struct{}

func (UnimplementedPingServiceServer) SendPing(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendPing not implemented")
}
func (UnimplementedPingServiceServer) ListenForMessages(*ListenRequest, grpc.ServerStreamingServer[Message]) error {
	return status.Errorf(codes.Unimplemented, "method ListenForMessages not implemented")
}
func (UnimplementedPingServiceServer) mustEmbedUnimplementedPingServiceServer() {}
func (UnimplementedPingServiceServer) testEmbeddedByValue()                     {}

// UnsafePingServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PingServiceServer will
// result in compilation errors.
type UnsafePingServiceServer interface {
	mustEmbedUnimplementedPingServiceServer()
}

func RegisterPingServiceServer(s grpc.ServiceRegistrar, srv PingServiceServer) {
	// If the following call pancis, it indicates UnimplementedPingServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&PingService_ServiceDesc, srv)
}

func _PingService_SendPing_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PingServiceServer).SendPing(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PingService_SendPing_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PingServiceServer).SendPing(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PingService_ListenForMessages_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListenRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PingServiceServer).ListenForMessages(m, &grpc.GenericServerStream[ListenRequest, Message]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type PingService_ListenForMessagesServer = grpc.ServerStreamingServer[Message]

// PingService_ServiceDesc is the grpc.ServiceDesc for PingService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PingService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ping.PingService",
	HandlerType: (*PingServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendPing",
			Handler:    _PingService_SendPing_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ListenForMessages",
			Handler:       _PingService_ListenForMessages_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/ping.proto",
}
