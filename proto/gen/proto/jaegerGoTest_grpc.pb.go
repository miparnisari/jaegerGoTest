// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: proto/jaegerGoTest.proto

package jaegerGoTest

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	JaegerGoTest_GetStoreID_FullMethodName         = "/jaegerGoTest.JaegerGoTest/GetStoreID"
	JaegerGoTest_StreamedGetStoreID_FullMethodName = "/jaegerGoTest.JaegerGoTest/StreamedGetStoreID"
)

// JaegerGoTestClient is the client API for JaegerGoTest service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type JaegerGoTestClient interface {
	GetStoreID(ctx context.Context, in *GetStoreRequest, opts ...grpc.CallOption) (*GetStoreResponse, error)
	StreamedGetStoreID(ctx context.Context, in *StreamedGetStoreRequest, opts ...grpc.CallOption) (JaegerGoTest_StreamedGetStoreIDClient, error)
}

type jaegerGoTestClient struct {
	cc grpc.ClientConnInterface
}

func NewJaegerGoTestClient(cc grpc.ClientConnInterface) JaegerGoTestClient {
	return &jaegerGoTestClient{cc}
}

func (c *jaegerGoTestClient) GetStoreID(ctx context.Context, in *GetStoreRequest, opts ...grpc.CallOption) (*GetStoreResponse, error) {
	out := new(GetStoreResponse)
	err := c.cc.Invoke(ctx, JaegerGoTest_GetStoreID_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *jaegerGoTestClient) StreamedGetStoreID(ctx context.Context, in *StreamedGetStoreRequest, opts ...grpc.CallOption) (JaegerGoTest_StreamedGetStoreIDClient, error) {
	stream, err := c.cc.NewStream(ctx, &JaegerGoTest_ServiceDesc.Streams[0], JaegerGoTest_StreamedGetStoreID_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &jaegerGoTestStreamedGetStoreIDClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type JaegerGoTest_StreamedGetStoreIDClient interface {
	Recv() (*StreamedGetStoreResponse, error)
	grpc.ClientStream
}

type jaegerGoTestStreamedGetStoreIDClient struct {
	grpc.ClientStream
}

func (x *jaegerGoTestStreamedGetStoreIDClient) Recv() (*StreamedGetStoreResponse, error) {
	m := new(StreamedGetStoreResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// JaegerGoTestServer is the server API for JaegerGoTest service.
// All implementations must embed UnimplementedJaegerGoTestServer
// for forward compatibility
type JaegerGoTestServer interface {
	GetStoreID(context.Context, *GetStoreRequest) (*GetStoreResponse, error)
	StreamedGetStoreID(*StreamedGetStoreRequest, JaegerGoTest_StreamedGetStoreIDServer) error
	mustEmbedUnimplementedJaegerGoTestServer()
}

// UnimplementedJaegerGoTestServer must be embedded to have forward compatible implementations.
type UnimplementedJaegerGoTestServer struct {
}

func (UnimplementedJaegerGoTestServer) GetStoreID(context.Context, *GetStoreRequest) (*GetStoreResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStoreID not implemented")
}
func (UnimplementedJaegerGoTestServer) StreamedGetStoreID(*StreamedGetStoreRequest, JaegerGoTest_StreamedGetStoreIDServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamedGetStoreID not implemented")
}
func (UnimplementedJaegerGoTestServer) mustEmbedUnimplementedJaegerGoTestServer() {}

// UnsafeJaegerGoTestServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to JaegerGoTestServer will
// result in compilation errors.
type UnsafeJaegerGoTestServer interface {
	mustEmbedUnimplementedJaegerGoTestServer()
}

func RegisterJaegerGoTestServer(s grpc.ServiceRegistrar, srv JaegerGoTestServer) {
	s.RegisterService(&JaegerGoTest_ServiceDesc, srv)
}

func _JaegerGoTest_GetStoreID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStoreRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(JaegerGoTestServer).GetStoreID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: JaegerGoTest_GetStoreID_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(JaegerGoTestServer).GetStoreID(ctx, req.(*GetStoreRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _JaegerGoTest_StreamedGetStoreID_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamedGetStoreRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(JaegerGoTestServer).StreamedGetStoreID(m, &jaegerGoTestStreamedGetStoreIDServer{stream})
}

type JaegerGoTest_StreamedGetStoreIDServer interface {
	Send(*StreamedGetStoreResponse) error
	grpc.ServerStream
}

type jaegerGoTestStreamedGetStoreIDServer struct {
	grpc.ServerStream
}

func (x *jaegerGoTestStreamedGetStoreIDServer) Send(m *StreamedGetStoreResponse) error {
	return x.ServerStream.SendMsg(m)
}

// JaegerGoTest_ServiceDesc is the grpc.ServiceDesc for JaegerGoTest service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var JaegerGoTest_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "jaegerGoTest.JaegerGoTest",
	HandlerType: (*JaegerGoTestServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetStoreID",
			Handler:    _JaegerGoTest_GetStoreID_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamedGetStoreID",
			Handler:       _JaegerGoTest_StreamedGetStoreID_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/jaegerGoTest.proto",
}
