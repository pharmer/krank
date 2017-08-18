// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ssh.proto

/*
Package v1beta1 is a generated protocol buffer package.

It is generated from these files:
	ssh.proto

It has these top-level messages:
	SSHGetRequest
	SSHGetResponse
	SSHKey
*/
package v1beta1

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "google.golang.org/genproto/googleapis/api/annotations"

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

// Use specific requests for protos
type SSHGetRequest struct {
	Namespace    string `protobuf:"bytes,1,opt,name=namespace" json:"namespace,omitempty"`
	ClusterName  string `protobuf:"bytes,2,opt,name=cluster_name,json=clusterName" json:"cluster_name,omitempty"`
	InstanceName string `protobuf:"bytes,3,opt,name=instance_name,json=instanceName" json:"instance_name,omitempty"`
}

func (m *SSHGetRequest) Reset()                    { *m = SSHGetRequest{} }
func (m *SSHGetRequest) String() string            { return proto.CompactTextString(m) }
func (*SSHGetRequest) ProtoMessage()               {}
func (*SSHGetRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *SSHGetRequest) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

func (m *SSHGetRequest) GetClusterName() string {
	if m != nil {
		return m.ClusterName
	}
	return ""
}

func (m *SSHGetRequest) GetInstanceName() string {
	if m != nil {
		return m.InstanceName
	}
	return ""
}

// return phid ?
type SSHGetResponse struct {
	SshKey       *SSHKey `protobuf:"bytes,1,opt,name=ssh_key,json=sshKey" json:"ssh_key,omitempty"`
	InstanceAddr string  `protobuf:"bytes,2,opt,name=instance_addr,json=instanceAddr" json:"instance_addr,omitempty"`
	InstancePort int32   `protobuf:"varint,3,opt,name=instance_port,json=instancePort" json:"instance_port,omitempty"`
	User         string  `protobuf:"bytes,4,opt,name=user" json:"user,omitempty"`
	Command      string  `protobuf:"bytes,5,opt,name=command" json:"command,omitempty"`
}

func (m *SSHGetResponse) Reset()                    { *m = SSHGetResponse{} }
func (m *SSHGetResponse) String() string            { return proto.CompactTextString(m) }
func (*SSHGetResponse) ProtoMessage()               {}
func (*SSHGetResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *SSHGetResponse) GetSshKey() *SSHKey {
	if m != nil {
		return m.SshKey
	}
	return nil
}

func (m *SSHGetResponse) GetInstanceAddr() string {
	if m != nil {
		return m.InstanceAddr
	}
	return ""
}

func (m *SSHGetResponse) GetInstancePort() int32 {
	if m != nil {
		return m.InstancePort
	}
	return 0
}

func (m *SSHGetResponse) GetUser() string {
	if m != nil {
		return m.User
	}
	return ""
}

func (m *SSHGetResponse) GetCommand() string {
	if m != nil {
		return m.Command
	}
	return ""
}

type SSHKey struct {
	PublicKey          []byte `protobuf:"bytes,1,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	PrivateKey         []byte `protobuf:"bytes,2,opt,name=private_key,json=privateKey,proto3" json:"private_key,omitempty"`
	AwsFingerprint     string `protobuf:"bytes,3,opt,name=aws_fingerprint,json=awsFingerprint" json:"aws_fingerprint,omitempty"`
	OpensshFingerprint string `protobuf:"bytes,4,opt,name=openssh_fingerprint,json=opensshFingerprint" json:"openssh_fingerprint,omitempty"`
}

func (m *SSHKey) Reset()                    { *m = SSHKey{} }
func (m *SSHKey) String() string            { return proto.CompactTextString(m) }
func (*SSHKey) ProtoMessage()               {}
func (*SSHKey) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *SSHKey) GetPublicKey() []byte {
	if m != nil {
		return m.PublicKey
	}
	return nil
}

func (m *SSHKey) GetPrivateKey() []byte {
	if m != nil {
		return m.PrivateKey
	}
	return nil
}

func (m *SSHKey) GetAwsFingerprint() string {
	if m != nil {
		return m.AwsFingerprint
	}
	return ""
}

func (m *SSHKey) GetOpensshFingerprint() string {
	if m != nil {
		return m.OpensshFingerprint
	}
	return ""
}

func init() {
	proto.RegisterType((*SSHGetRequest)(nil), "appscode.ssh.v1beta1.SSHGetRequest")
	proto.RegisterType((*SSHGetResponse)(nil), "appscode.ssh.v1beta1.SSHGetResponse")
	proto.RegisterType((*SSHKey)(nil), "appscode.ssh.v1beta1.SSHKey")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for SSH service

type SSHClient interface {
	Get(ctx context.Context, in *SSHGetRequest, opts ...grpc.CallOption) (*SSHGetResponse, error)
}

type sSHClient struct {
	cc *grpc.ClientConn
}

func NewSSHClient(cc *grpc.ClientConn) SSHClient {
	return &sSHClient{cc}
}

func (c *sSHClient) Get(ctx context.Context, in *SSHGetRequest, opts ...grpc.CallOption) (*SSHGetResponse, error) {
	out := new(SSHGetResponse)
	err := grpc.Invoke(ctx, "/appscode.ssh.v1beta1.SSH/Get", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for SSH service

type SSHServer interface {
	Get(context.Context, *SSHGetRequest) (*SSHGetResponse, error)
}

func RegisterSSHServer(s *grpc.Server, srv SSHServer) {
	s.RegisterService(&_SSH_serviceDesc, srv)
}

func _SSH_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SSHGetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SSHServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/appscode.ssh.v1beta1.SSH/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SSHServer).Get(ctx, req.(*SSHGetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _SSH_serviceDesc = grpc.ServiceDesc{
	ServiceName: "appscode.ssh.v1beta1.SSH",
	HandlerType: (*SSHServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _SSH_Get_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ssh.proto",
}

func init() { proto.RegisterFile("ssh.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 431 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x92, 0x4d, 0x6e, 0xd3, 0x40,
	0x14, 0x80, 0xe5, 0xa4, 0x4d, 0xc8, 0x4b, 0x5a, 0xa4, 0x81, 0x85, 0x55, 0x05, 0x01, 0x09, 0x08,
	0x24, 0xa4, 0x58, 0x6d, 0xc5, 0x01, 0xe8, 0x82, 0x46, 0xaa, 0x84, 0x22, 0xcf, 0x8e, 0x8d, 0x35,
	0xb1, 0x1f, 0xf1, 0x40, 0x3c, 0x33, 0xcc, 0x1b, 0xb7, 0xca, 0x82, 0x0d, 0xe2, 0x06, 0xdc, 0x80,
	0x8b, 0x70, 0x08, 0xae, 0xc0, 0x41, 0x90, 0xc7, 0x76, 0x9d, 0x4a, 0xc0, 0x6e, 0xf2, 0xbd, 0xef,
	0xfd, 0xe4, 0xf9, 0xc1, 0x88, 0x28, 0x5f, 0x18, 0xab, 0x9d, 0x66, 0x0f, 0x85, 0x31, 0x94, 0xea,
	0x0c, 0x17, 0x15, 0xbb, 0x3e, 0x5d, 0xa3, 0x13, 0xa7, 0x27, 0xd3, 0x8d, 0xd6, 0x9b, 0x2d, 0x46,
	0xc2, 0xc8, 0x48, 0x28, 0xa5, 0x9d, 0x70, 0x52, 0x2b, 0xaa, 0x73, 0x66, 0x25, 0x1c, 0x71, 0xbe,
	0xbc, 0x44, 0x17, 0xe3, 0xe7, 0x12, 0xc9, 0xb1, 0x29, 0x8c, 0x94, 0x28, 0x90, 0x8c, 0x48, 0x31,
	0x0c, 0x9e, 0x04, 0x2f, 0x47, 0x71, 0x07, 0xd8, 0x53, 0x98, 0xa4, 0xdb, 0x92, 0x1c, 0xda, 0xa4,
	0x82, 0x61, 0xcf, 0x0b, 0xe3, 0x86, 0xbd, 0x13, 0x05, 0xb2, 0x39, 0x1c, 0x49, 0x45, 0x4e, 0xa8,
	0x14, 0x6b, 0xa7, 0xef, 0x9d, 0x49, 0x0b, 0x2b, 0x69, 0xf6, 0x33, 0x80, 0xe3, 0xb6, 0x2f, 0x19,
	0xad, 0x08, 0xd9, 0x6b, 0x18, 0x12, 0xe5, 0xc9, 0x27, 0xdc, 0xf9, 0xb6, 0xe3, 0xb3, 0xe9, 0xe2,
	0x6f, 0xff, 0x67, 0xc1, 0xf9, 0xf2, 0x0a, 0x77, 0xf1, 0x80, 0x28, 0xbf, 0xc2, 0xdd, 0x9d, 0x76,
	0x22, 0xcb, 0x6c, 0x33, 0xd2, 0x6d, 0xbb, 0x37, 0x59, 0x66, 0xef, 0x48, 0x46, 0x5b, 0xe7, 0x67,
	0x3a, 0xec, 0xa4, 0x95, 0xb6, 0x8e, 0x31, 0x38, 0x28, 0x09, 0x6d, 0x78, 0xe0, 0x0b, 0xf8, 0x37,
	0x0b, 0x61, 0x98, 0xea, 0xa2, 0x10, 0x2a, 0x0b, 0x0f, 0x3d, 0x6e, 0x7f, 0xce, 0x7e, 0x04, 0x30,
	0xa8, 0x47, 0x61, 0x8f, 0x00, 0x4c, 0xb9, 0xde, 0xca, 0xf4, 0x76, 0xf8, 0x49, 0x3c, 0xaa, 0x49,
	0x15, 0x7e, 0x0c, 0x63, 0x63, 0xe5, 0xb5, 0x70, 0xe8, 0xe3, 0x3d, 0x1f, 0x87, 0x06, 0x55, 0xc2,
	0x0b, 0xb8, 0x2f, 0x6e, 0x28, 0xf9, 0x20, 0xd5, 0x06, 0xad, 0xb1, 0x52, 0xb9, 0x66, 0x67, 0xc7,
	0xe2, 0x86, 0xde, 0x76, 0x94, 0x45, 0xf0, 0x40, 0x1b, 0x54, 0xd5, 0x9a, 0xf6, 0xe5, 0x7a, 0x60,
	0xd6, 0x84, 0xf6, 0x12, 0xce, 0xbe, 0x05, 0xd0, 0xe7, 0x7c, 0xc9, 0xbe, 0x40, 0xff, 0x12, 0x1d,
	0x9b, 0xff, 0x73, 0xa3, 0xdd, 0x01, 0x9c, 0x3c, 0xfb, 0xbf, 0x54, 0x7f, 0xad, 0xd9, 0xab, 0xaf,
	0xbf, 0x7e, 0x7f, 0xef, 0x3d, 0x67, 0xf3, 0x28, 0x69, 0x75, 0x7f, 0x61, 0x44, 0x79, 0xd4, 0xa4,
	0xf8, 0xf7, 0x47, 0xd2, 0xea, 0xe2, 0x1c, 0xa6, 0xa9, 0x2e, 0xba, 0xba, 0xc2, 0xc8, 0xfd, 0xda,
	0x17, 0xf7, 0x38, 0x5f, 0xae, 0xaa, 0x73, 0x5c, 0x05, 0xef, 0x87, 0x0d, 0x5c, 0x0f, 0xfc, 0x81,
	0x9e, 0xff, 0x09, 0x00, 0x00, 0xff, 0xff, 0x6a, 0x15, 0xe6, 0xad, 0xe1, 0x02, 0x00, 0x00,
}