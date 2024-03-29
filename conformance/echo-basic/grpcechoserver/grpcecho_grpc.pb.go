// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.22.2
// source: grpcecho.proto

package grpcechoserver

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

// GrpcEchoClient is the client API for GrpcEcho service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GrpcEchoClient interface {
	Echo(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (*EchoResponse, error)
	// Behaves identically to Echo, but lives at a different method to
	// emulate the service having more than one method.
	EchoTwo(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (*EchoResponse, error)
	// An intentionally unimplemented method.
	EchoThree(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (*EchoResponse, error)
}

type grpcEchoClient struct {
	cc grpc.ClientConnInterface
}

func NewGrpcEchoClient(cc grpc.ClientConnInterface) GrpcEchoClient {
	return &grpcEchoClient{cc}
}

func (c *grpcEchoClient) Echo(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (*EchoResponse, error) {
	out := new(EchoResponse)
	err := c.cc.Invoke(ctx, "/gateway_api_conformance.echo_basic.grpcecho.GrpcEcho/Echo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcEchoClient) EchoTwo(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (*EchoResponse, error) {
	out := new(EchoResponse)
	err := c.cc.Invoke(ctx, "/gateway_api_conformance.echo_basic.grpcecho.GrpcEcho/EchoTwo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcEchoClient) EchoThree(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (*EchoResponse, error) {
	out := new(EchoResponse)
	err := c.cc.Invoke(ctx, "/gateway_api_conformance.echo_basic.grpcecho.GrpcEcho/EchoThree", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GrpcEchoServer is the server API for GrpcEcho service.
// All implementations must embed UnimplementedGrpcEchoServer
// for forward compatibility
type GrpcEchoServer interface {
	Echo(context.Context, *EchoRequest) (*EchoResponse, error)
	// Behaves identically to Echo, but lives at a different method to
	// emulate the service having more than one method.
	EchoTwo(context.Context, *EchoRequest) (*EchoResponse, error)
	// An intentionally unimplemented method.
	EchoThree(context.Context, *EchoRequest) (*EchoResponse, error)
	mustEmbedUnimplementedGrpcEchoServer()
}

// UnimplementedGrpcEchoServer must be embedded to have forward compatible implementations.
type UnimplementedGrpcEchoServer struct {
}

func (UnimplementedGrpcEchoServer) Echo(context.Context, *EchoRequest) (*EchoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Echo not implemented")
}
func (UnimplementedGrpcEchoServer) EchoTwo(context.Context, *EchoRequest) (*EchoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EchoTwo not implemented")
}
func (UnimplementedGrpcEchoServer) EchoThree(context.Context, *EchoRequest) (*EchoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EchoThree not implemented")
}
func (UnimplementedGrpcEchoServer) mustEmbedUnimplementedGrpcEchoServer() {}

// UnsafeGrpcEchoServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GrpcEchoServer will
// result in compilation errors.
type UnsafeGrpcEchoServer interface {
	mustEmbedUnimplementedGrpcEchoServer()
}

func RegisterGrpcEchoServer(s grpc.ServiceRegistrar, srv GrpcEchoServer) {
	s.RegisterService(&GrpcEcho_ServiceDesc, srv)
}

func _GrpcEcho_Echo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EchoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcEchoServer).Echo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gateway_api_conformance.echo_basic.grpcecho.GrpcEcho/Echo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcEchoServer).Echo(ctx, req.(*EchoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcEcho_EchoTwo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EchoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcEchoServer).EchoTwo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gateway_api_conformance.echo_basic.grpcecho.GrpcEcho/EchoTwo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcEchoServer).EchoTwo(ctx, req.(*EchoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcEcho_EchoThree_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EchoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcEchoServer).EchoThree(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gateway_api_conformance.echo_basic.grpcecho.GrpcEcho/EchoThree",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcEchoServer).EchoThree(ctx, req.(*EchoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GrpcEcho_ServiceDesc is the grpc.ServiceDesc for GrpcEcho service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GrpcEcho_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gateway_api_conformance.echo_basic.grpcecho.GrpcEcho",
	HandlerType: (*GrpcEchoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Echo",
			Handler:    _GrpcEcho_Echo_Handler,
		},
		{
			MethodName: "EchoTwo",
			Handler:    _GrpcEcho_EchoTwo_Handler,
		},
		{
			MethodName: "EchoThree",
			Handler:    _GrpcEcho_EchoThree_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpcecho.proto",
}
