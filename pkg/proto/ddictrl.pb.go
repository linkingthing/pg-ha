// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ddictrl.proto

package proto

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type DDICtrlResponse struct {
	Succeed              bool     `protobuf:"varint,1,opt,name=succeed,proto3" json:"succeed,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DDICtrlResponse) Reset()         { *m = DDICtrlResponse{} }
func (m *DDICtrlResponse) String() string { return proto.CompactTextString(m) }
func (*DDICtrlResponse) ProtoMessage()    {}
func (*DDICtrlResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ddictrl_32c66f5297d6edcc, []int{0}
}
func (m *DDICtrlResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DDICtrlResponse.Unmarshal(m, b)
}
func (m *DDICtrlResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DDICtrlResponse.Marshal(b, m, deterministic)
}
func (dst *DDICtrlResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DDICtrlResponse.Merge(dst, src)
}
func (m *DDICtrlResponse) XXX_Size() int {
	return xxx_messageInfo_DDICtrlResponse.Size(m)
}
func (m *DDICtrlResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DDICtrlResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DDICtrlResponse proto.InternalMessageInfo

func (m *DDICtrlResponse) GetSucceed() bool {
	if m != nil {
		return m.Succeed
	}
	return false
}

type DDICtrlRequest struct {
	MasterIp             string   `protobuf:"bytes,1,opt,name=master_ip,json=masterIp,proto3" json:"master_ip,omitempty"`
	SlaveIp              string   `protobuf:"bytes,2,opt,name=slave_ip,json=slaveIp,proto3" json:"slave_ip,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DDICtrlRequest) Reset()         { *m = DDICtrlRequest{} }
func (m *DDICtrlRequest) String() string { return proto.CompactTextString(m) }
func (*DDICtrlRequest) ProtoMessage()    {}
func (*DDICtrlRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ddictrl_32c66f5297d6edcc, []int{1}
}
func (m *DDICtrlRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DDICtrlRequest.Unmarshal(m, b)
}
func (m *DDICtrlRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DDICtrlRequest.Marshal(b, m, deterministic)
}
func (dst *DDICtrlRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DDICtrlRequest.Merge(dst, src)
}
func (m *DDICtrlRequest) XXX_Size() int {
	return xxx_messageInfo_DDICtrlRequest.Size(m)
}
func (m *DDICtrlRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DDICtrlRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DDICtrlRequest proto.InternalMessageInfo

func (m *DDICtrlRequest) GetMasterIp() string {
	if m != nil {
		return m.MasterIp
	}
	return ""
}

func (m *DDICtrlRequest) GetSlaveIp() string {
	if m != nil {
		return m.SlaveIp
	}
	return ""
}

func init() {
	proto.RegisterType((*DDICtrlResponse)(nil), "proto.DDICtrlResponse")
	proto.RegisterType((*DDICtrlRequest)(nil), "proto.DDICtrlRequest")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// DDICtrlManagerClient is the client API for DDICtrlManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type DDICtrlManagerClient interface {
	MasterUp(ctx context.Context, in *DDICtrlRequest, opts ...grpc.CallOption) (*DDICtrlResponse, error)
	MasterDown(ctx context.Context, in *DDICtrlRequest, opts ...grpc.CallOption) (*DDICtrlResponse, error)
}

type dDICtrlManagerClient struct {
	cc *grpc.ClientConn
}

func NewDDICtrlManagerClient(cc *grpc.ClientConn) DDICtrlManagerClient {
	return &dDICtrlManagerClient{cc}
}

func (c *dDICtrlManagerClient) MasterUp(ctx context.Context, in *DDICtrlRequest, opts ...grpc.CallOption) (*DDICtrlResponse, error) {
	out := new(DDICtrlResponse)
	err := c.cc.Invoke(ctx, "/proto.DDICtrlManager/MasterUp", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dDICtrlManagerClient) MasterDown(ctx context.Context, in *DDICtrlRequest, opts ...grpc.CallOption) (*DDICtrlResponse, error) {
	out := new(DDICtrlResponse)
	err := c.cc.Invoke(ctx, "/proto.DDICtrlManager/MasterDown", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DDICtrlManagerServer is the server API for DDICtrlManager service.
type DDICtrlManagerServer interface {
	MasterUp(context.Context, *DDICtrlRequest) (*DDICtrlResponse, error)
	MasterDown(context.Context, *DDICtrlRequest) (*DDICtrlResponse, error)
}

func RegisterDDICtrlManagerServer(s *grpc.Server, srv DDICtrlManagerServer) {
	s.RegisterService(&_DDICtrlManager_serviceDesc, srv)
}

func _DDICtrlManager_MasterUp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DDICtrlRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DDICtrlManagerServer).MasterUp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DDICtrlManager/MasterUp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DDICtrlManagerServer).MasterUp(ctx, req.(*DDICtrlRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DDICtrlManager_MasterDown_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DDICtrlRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DDICtrlManagerServer).MasterDown(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DDICtrlManager/MasterDown",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DDICtrlManagerServer).MasterDown(ctx, req.(*DDICtrlRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _DDICtrlManager_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.DDICtrlManager",
	HandlerType: (*DDICtrlManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MasterUp",
			Handler:    _DDICtrlManager_MasterUp_Handler,
		},
		{
			MethodName: "MasterDown",
			Handler:    _DDICtrlManager_MasterDown_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ddictrl.proto",
}

func init() { proto.RegisterFile("ddictrl.proto", fileDescriptor_ddictrl_32c66f5297d6edcc) }

var fileDescriptor_ddictrl_32c66f5297d6edcc = []byte{
	// 187 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4d, 0x49, 0xc9, 0x4c,
	0x2e, 0x29, 0xca, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x05, 0x53, 0x4a, 0xda, 0x5c,
	0xfc, 0x2e, 0x2e, 0x9e, 0xce, 0x25, 0x45, 0x39, 0x41, 0xa9, 0xc5, 0x05, 0xf9, 0x79, 0xc5, 0xa9,
	0x42, 0x12, 0x5c, 0xec, 0xc5, 0xa5, 0xc9, 0xc9, 0xa9, 0xa9, 0x29, 0x12, 0x8c, 0x0a, 0x8c, 0x1a,
	0x1c, 0x41, 0x30, 0xae, 0x92, 0x07, 0x17, 0x1f, 0x5c, 0x71, 0x61, 0x69, 0x6a, 0x71, 0x89, 0x90,
	0x34, 0x17, 0x67, 0x6e, 0x62, 0x71, 0x49, 0x6a, 0x51, 0x7c, 0x66, 0x01, 0x58, 0x35, 0x67, 0x10,
	0x07, 0x44, 0xc0, 0xb3, 0x40, 0x48, 0x92, 0x8b, 0xa3, 0x38, 0x27, 0xb1, 0x2c, 0x15, 0x24, 0xc7,
	0x04, 0x96, 0x63, 0x07, 0xf3, 0x3d, 0x0b, 0x8c, 0x7a, 0x18, 0xe1, 0x46, 0xf9, 0x26, 0xe6, 0x25,
	0xa6, 0xa7, 0x16, 0x09, 0x59, 0x73, 0x71, 0xf8, 0x82, 0x75, 0x86, 0x16, 0x08, 0x89, 0x42, 0x1c,
	0xa9, 0x87, 0x6a, 0x9b, 0x94, 0x18, 0xba, 0x30, 0xc4, 0xc5, 0x4a, 0x0c, 0x42, 0xb6, 0x5c, 0x5c,
	0x10, 0xcd, 0x2e, 0xf9, 0xe5, 0x79, 0x24, 0x6b, 0x4f, 0x62, 0x03, 0x4b, 0x18, 0x03, 0x02, 0x00,
	0x00, 0xff, 0xff, 0x6b, 0x54, 0x88, 0x62, 0x24, 0x01, 0x00, 0x00,
}
