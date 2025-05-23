// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: careerup/v1/ilo.proto

package careerupv1

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
	IloService_SubmitIloTestResult_FullMethodName     = "/careerup.v1.IloService/SubmitIloTestResult"
	IloService_GetIloTestResults_FullMethodName       = "/careerup.v1.IloService/GetIloTestResults"
	IloService_GetIloTestResult_FullMethodName        = "/careerup.v1.IloService/GetIloTestResult"
	IloService_GetIloTest_FullMethodName              = "/careerup.v1.IloService/GetIloTest"
	IloService_GetIloCareerSuggestions_FullMethodName = "/careerup.v1.IloService/GetIloCareerSuggestions"
)

// IloServiceClient is the client API for IloService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IloServiceClient interface {
	// Submit a completed ILO test
	SubmitIloTestResult(ctx context.Context, in *SubmitIloTestResultRequest, opts ...grpc.CallOption) (*SubmitIloTestResultResponse, error)
	// Get all ILO test results for a user
	GetIloTestResults(ctx context.Context, in *GetIloTestResultsRequest, opts ...grpc.CallOption) (*GetIloTestResultsResponse, error)
	// Get a specific ILO test result by ID
	GetIloTestResult(ctx context.Context, in *GetIloTestResultRequest, opts ...grpc.CallOption) (*GetIloTestResultResponse, error)
	// Get ILO test questions and structure
	GetIloTest(ctx context.Context, in *GetIloTestRequest, opts ...grpc.CallOption) (*GetIloTestResponse, error)
	// Get career suggestions based on domain scores
	GetIloCareerSuggestions(ctx context.Context, in *GetIloCareerSuggestionsRequest, opts ...grpc.CallOption) (*GetIloCareerSuggestionsResponse, error)
}

type iloServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewIloServiceClient(cc grpc.ClientConnInterface) IloServiceClient {
	return &iloServiceClient{cc}
}

func (c *iloServiceClient) SubmitIloTestResult(ctx context.Context, in *SubmitIloTestResultRequest, opts ...grpc.CallOption) (*SubmitIloTestResultResponse, error) {
	out := new(SubmitIloTestResultResponse)
	err := c.cc.Invoke(ctx, IloService_SubmitIloTestResult_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iloServiceClient) GetIloTestResults(ctx context.Context, in *GetIloTestResultsRequest, opts ...grpc.CallOption) (*GetIloTestResultsResponse, error) {
	out := new(GetIloTestResultsResponse)
	err := c.cc.Invoke(ctx, IloService_GetIloTestResults_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iloServiceClient) GetIloTestResult(ctx context.Context, in *GetIloTestResultRequest, opts ...grpc.CallOption) (*GetIloTestResultResponse, error) {
	out := new(GetIloTestResultResponse)
	err := c.cc.Invoke(ctx, IloService_GetIloTestResult_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iloServiceClient) GetIloTest(ctx context.Context, in *GetIloTestRequest, opts ...grpc.CallOption) (*GetIloTestResponse, error) {
	out := new(GetIloTestResponse)
	err := c.cc.Invoke(ctx, IloService_GetIloTest_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iloServiceClient) GetIloCareerSuggestions(ctx context.Context, in *GetIloCareerSuggestionsRequest, opts ...grpc.CallOption) (*GetIloCareerSuggestionsResponse, error) {
	out := new(GetIloCareerSuggestionsResponse)
	err := c.cc.Invoke(ctx, IloService_GetIloCareerSuggestions_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IloServiceServer is the server API for IloService service.
// All implementations must embed UnimplementedIloServiceServer
// for forward compatibility
type IloServiceServer interface {
	// Submit a completed ILO test
	SubmitIloTestResult(context.Context, *SubmitIloTestResultRequest) (*SubmitIloTestResultResponse, error)
	// Get all ILO test results for a user
	GetIloTestResults(context.Context, *GetIloTestResultsRequest) (*GetIloTestResultsResponse, error)
	// Get a specific ILO test result by ID
	GetIloTestResult(context.Context, *GetIloTestResultRequest) (*GetIloTestResultResponse, error)
	// Get ILO test questions and structure
	GetIloTest(context.Context, *GetIloTestRequest) (*GetIloTestResponse, error)
	// Get career suggestions based on domain scores
	GetIloCareerSuggestions(context.Context, *GetIloCareerSuggestionsRequest) (*GetIloCareerSuggestionsResponse, error)
	mustEmbedUnimplementedIloServiceServer()
}

// UnimplementedIloServiceServer must be embedded to have forward compatible implementations.
type UnimplementedIloServiceServer struct {
}

func (UnimplementedIloServiceServer) SubmitIloTestResult(context.Context, *SubmitIloTestResultRequest) (*SubmitIloTestResultResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitIloTestResult not implemented")
}
func (UnimplementedIloServiceServer) GetIloTestResults(context.Context, *GetIloTestResultsRequest) (*GetIloTestResultsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIloTestResults not implemented")
}
func (UnimplementedIloServiceServer) GetIloTestResult(context.Context, *GetIloTestResultRequest) (*GetIloTestResultResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIloTestResult not implemented")
}
func (UnimplementedIloServiceServer) GetIloTest(context.Context, *GetIloTestRequest) (*GetIloTestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIloTest not implemented")
}
func (UnimplementedIloServiceServer) GetIloCareerSuggestions(context.Context, *GetIloCareerSuggestionsRequest) (*GetIloCareerSuggestionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIloCareerSuggestions not implemented")
}
func (UnimplementedIloServiceServer) mustEmbedUnimplementedIloServiceServer() {}

// UnsafeIloServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IloServiceServer will
// result in compilation errors.
type UnsafeIloServiceServer interface {
	mustEmbedUnimplementedIloServiceServer()
}

func RegisterIloServiceServer(s grpc.ServiceRegistrar, srv IloServiceServer) {
	s.RegisterService(&IloService_ServiceDesc, srv)
}

func _IloService_SubmitIloTestResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitIloTestResultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IloServiceServer).SubmitIloTestResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IloService_SubmitIloTestResult_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IloServiceServer).SubmitIloTestResult(ctx, req.(*SubmitIloTestResultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IloService_GetIloTestResults_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetIloTestResultsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IloServiceServer).GetIloTestResults(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IloService_GetIloTestResults_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IloServiceServer).GetIloTestResults(ctx, req.(*GetIloTestResultsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IloService_GetIloTestResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetIloTestResultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IloServiceServer).GetIloTestResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IloService_GetIloTestResult_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IloServiceServer).GetIloTestResult(ctx, req.(*GetIloTestResultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IloService_GetIloTest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetIloTestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IloServiceServer).GetIloTest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IloService_GetIloTest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IloServiceServer).GetIloTest(ctx, req.(*GetIloTestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IloService_GetIloCareerSuggestions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetIloCareerSuggestionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IloServiceServer).GetIloCareerSuggestions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IloService_GetIloCareerSuggestions_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IloServiceServer).GetIloCareerSuggestions(ctx, req.(*GetIloCareerSuggestionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// IloService_ServiceDesc is the grpc.ServiceDesc for IloService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var IloService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "careerup.v1.IloService",
	HandlerType: (*IloServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitIloTestResult",
			Handler:    _IloService_SubmitIloTestResult_Handler,
		},
		{
			MethodName: "GetIloTestResults",
			Handler:    _IloService_GetIloTestResults_Handler,
		},
		{
			MethodName: "GetIloTestResult",
			Handler:    _IloService_GetIloTestResult_Handler,
		},
		{
			MethodName: "GetIloTest",
			Handler:    _IloService_GetIloTest_Handler,
		},
		{
			MethodName: "GetIloCareerSuggestions",
			Handler:    _IloService_GetIloCareerSuggestions_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "careerup/v1/ilo.proto",
}
