// Code generated by protoc-gen-go. DO NOT EDIT.
// source: build.proto

package build

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Request struct {
	Project              string   `protobuf:"bytes,1,opt,name=project,proto3" json:"project,omitempty"`
	Branch               string   `protobuf:"bytes,2,opt,name=branch,proto3" json:"branch,omitempty"`
	Env                  string   `protobuf:"bytes,3,opt,name=env,proto3" json:"env,omitempty"`
	Commitid             string   `protobuf:"bytes,4,opt,name=commitid,proto3" json:"commitid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_14ce178a580e4ede, []int{0}
}

func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (m *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(m, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

func (m *Request) GetProject() string {
	if m != nil {
		return m.Project
	}
	return ""
}

func (m *Request) GetBranch() string {
	if m != nil {
		return m.Branch
	}
	return ""
}

func (m *Request) GetEnv() string {
	if m != nil {
		return m.Env
	}
	return ""
}

func (m *Request) GetCommitid() string {
	if m != nil {
		return m.Commitid
	}
	return ""
}

type Response struct {
	Output               string   `protobuf:"bytes,1,opt,name=output,proto3" json:"output,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}
func (*Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_14ce178a580e4ede, []int{1}
}

func (m *Response) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Response.Unmarshal(m, b)
}
func (m *Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Response.Marshal(b, m, deterministic)
}
func (m *Response) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Response.Merge(m, src)
}
func (m *Response) XXX_Size() int {
	return xxx_messageInfo_Response.Size(m)
}
func (m *Response) XXX_DiscardUnknown() {
	xxx_messageInfo_Response.DiscardUnknown(m)
}

var xxx_messageInfo_Response proto.InternalMessageInfo

func (m *Response) GetOutput() string {
	if m != nil {
		return m.Output
	}
	return ""
}

func init() {
	proto.RegisterType((*Request)(nil), "Request")
	proto.RegisterType((*Response)(nil), "Response")
}

func init() { proto.RegisterFile("build.proto", fileDescriptor_14ce178a580e4ede) }

var fileDescriptor_14ce178a580e4ede = []byte{
	// 171 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0x8f, 0xc1, 0x0e, 0x82, 0x30,
	0x10, 0x44, 0x45, 0x14, 0xca, 0x7a, 0x31, 0x7b, 0x30, 0x0d, 0x27, 0xd2, 0x93, 0x07, 0x43, 0x8c,
	0xfe, 0x81, 0x9f, 0xc0, 0x1f, 0x48, 0xd9, 0xc4, 0x1a, 0x61, 0x2b, 0x6d, 0xf9, 0x7e, 0x43, 0x03,
	0xde, 0xe6, 0xcd, 0x26, 0xb3, 0x33, 0x70, 0x68, 0x83, 0xf9, 0x74, 0xb5, 0x1d, 0xd9, 0xb3, 0x32,
	0x90, 0x37, 0xf4, 0x0d, 0xe4, 0x3c, 0x4a, 0xc8, 0xed, 0xc8, 0x6f, 0xd2, 0x5e, 0x26, 0x55, 0x72,
	0x2e, 0x9a, 0x15, 0xf1, 0x04, 0x59, 0x3b, 0x3e, 0x07, 0xfd, 0x92, 0xdb, 0x78, 0x58, 0x08, 0x8f,
	0x90, 0xd2, 0x30, 0xc9, 0x34, 0x9a, 0xb3, 0xc4, 0x12, 0x84, 0xe6, 0xbe, 0x37, 0xde, 0x74, 0x72,
	0x17, 0xed, 0x3f, 0x2b, 0x05, 0xa2, 0x21, 0x67, 0x79, 0x70, 0x34, 0x27, 0x72, 0xf0, 0x36, 0xac,
	0xaf, 0x16, 0xba, 0x5d, 0x40, 0x3c, 0xe6, 0x76, 0x6e, 0xd2, 0x58, 0xc1, 0x3e, 0x6a, 0x14, 0xf5,
	0x52, 0xb1, 0x2c, 0xea, 0x35, 0x41, 0x6d, 0xae, 0x49, 0x9b, 0xc5, 0x0d, 0xf7, 0x5f, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x3d, 0xb1, 0x4f, 0xf8, 0xd2, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// BuildsvcClient is the client API for Buildsvc service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type BuildsvcClient interface {
	Build(ctx context.Context, in *Request, opts ...grpc.CallOption) (Buildsvc_BuildClient, error)
}

type buildsvcClient struct {
	cc *grpc.ClientConn
}

func NewBuildsvcClient(cc *grpc.ClientConn) BuildsvcClient {
	return &buildsvcClient{cc}
}

func (c *buildsvcClient) Build(ctx context.Context, in *Request, opts ...grpc.CallOption) (Buildsvc_BuildClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Buildsvc_serviceDesc.Streams[0], "/Buildsvc/Build", opts...)
	if err != nil {
		return nil, err
	}
	x := &buildsvcBuildClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Buildsvc_BuildClient interface {
	Recv() (*Response, error)
	grpc.ClientStream
}

type buildsvcBuildClient struct {
	grpc.ClientStream
}

func (x *buildsvcBuildClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// BuildsvcServer is the server API for Buildsvc service.
type BuildsvcServer interface {
	Build(*Request, Buildsvc_BuildServer) error
}

func RegisterBuildsvcServer(s *grpc.Server, srv BuildsvcServer) {
	s.RegisterService(&_Buildsvc_serviceDesc, srv)
}

func _Buildsvc_Build_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Request)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BuildsvcServer).Build(m, &buildsvcBuildServer{stream})
}

type Buildsvc_BuildServer interface {
	Send(*Response) error
	grpc.ServerStream
}

type buildsvcBuildServer struct {
	grpc.ServerStream
}

func (x *buildsvcBuildServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

var _Buildsvc_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Buildsvc",
	HandlerType: (*BuildsvcServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Build",
			Handler:       _Buildsvc_Build_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "build.proto",
}
