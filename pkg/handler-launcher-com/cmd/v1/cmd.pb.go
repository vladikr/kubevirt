// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pkg/handler-launcher-com/cmd/v1/cmd.proto

/*
Package v1 is a generated protocol buffer package.

It is generated from these files:
	pkg/handler-launcher-com/cmd/v1/cmd.proto

It has these top-level messages:
	VMI
	SMBios
	HostProcessorsMap
	VirtualMachineOptions
	VMIRequest
	MigrationRequest
	EmptyRequest
	Response
	DomainResponse
	DomainStatsResponse
*/
package v1

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

type VMI struct {
	VmiJson []byte `protobuf:"bytes,1,opt,name=vmiJson,proto3" json:"vmiJson,omitempty"`
}

func (m *VMI) Reset()                    { *m = VMI{} }
func (m *VMI) String() string            { return proto.CompactTextString(m) }
func (*VMI) ProtoMessage()               {}
func (*VMI) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *VMI) GetVmiJson() []byte {
	if m != nil {
		return m.VmiJson
	}
	return nil
}

type SMBios struct {
	Manufacturer string `protobuf:"bytes,1,opt,name=manufacturer" json:"manufacturer,omitempty"`
	Product      string `protobuf:"bytes,2,opt,name=product" json:"product,omitempty"`
	Version      string `protobuf:"bytes,3,opt,name=version" json:"version,omitempty"`
	Sku          string `protobuf:"bytes,4,opt,name=sku" json:"sku,omitempty"`
	Family       string `protobuf:"bytes,5,opt,name=family" json:"family,omitempty"`
}

func (m *SMBios) Reset()                    { *m = SMBios{} }
func (m *SMBios) String() string            { return proto.CompactTextString(m) }
func (*SMBios) ProtoMessage()               {}
func (*SMBios) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *SMBios) GetManufacturer() string {
	if m != nil {
		return m.Manufacturer
	}
	return ""
}

func (m *SMBios) GetProduct() string {
	if m != nil {
		return m.Product
	}
	return ""
}

func (m *SMBios) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *SMBios) GetSku() string {
	if m != nil {
		return m.Sku
	}
	return ""
}

func (m *SMBios) GetFamily() string {
	if m != nil {
		return m.Family
	}
	return ""
}

type HostProcessorsMap struct {
	ProcMapJson []byte `protobuf:"bytes,1,opt,name=procMapJson,proto3" json:"procMapJson,omitempty"`
}

func (m *HostProcessorsMap) Reset()                    { *m = HostProcessorsMap{} }
func (m *HostProcessorsMap) String() string            { return proto.CompactTextString(m) }
func (*HostProcessorsMap) ProtoMessage()               {}
func (*HostProcessorsMap) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *HostProcessorsMap) GetProcMapJson() []byte {
	if m != nil {
		return m.ProcMapJson
	}
	return nil
}

type VirtualMachineOptions struct {
	VirtualMachineSMBios *SMBios            `protobuf:"bytes,1,opt,name=VirtualMachineSMBios" json:"VirtualMachineSMBios,omitempty"`
	ProcessorsMap        *HostProcessorsMap `protobuf:"bytes,2,opt,name=ProcessorsMap" json:"ProcessorsMap,omitempty"`
}

func (m *VirtualMachineOptions) Reset()                    { *m = VirtualMachineOptions{} }
func (m *VirtualMachineOptions) String() string            { return proto.CompactTextString(m) }
func (*VirtualMachineOptions) ProtoMessage()               {}
func (*VirtualMachineOptions) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *VirtualMachineOptions) GetVirtualMachineSMBios() *SMBios {
	if m != nil {
		return m.VirtualMachineSMBios
	}
	return nil
}

func (m *VirtualMachineOptions) GetProcessorsMap() *HostProcessorsMap {
	if m != nil {
		return m.ProcessorsMap
	}
	return nil
}

type VMIRequest struct {
	Vmi     *VMI                   `protobuf:"bytes,1,opt,name=vmi" json:"vmi,omitempty"`
	Options *VirtualMachineOptions `protobuf:"bytes,2,opt,name=options" json:"options,omitempty"`
}

func (m *VMIRequest) Reset()                    { *m = VMIRequest{} }
func (m *VMIRequest) String() string            { return proto.CompactTextString(m) }
func (*VMIRequest) ProtoMessage()               {}
func (*VMIRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *VMIRequest) GetVmi() *VMI {
	if m != nil {
		return m.Vmi
	}
	return nil
}

func (m *VMIRequest) GetOptions() *VirtualMachineOptions {
	if m != nil {
		return m.Options
	}
	return nil
}

type MigrationRequest struct {
	Vmi     *VMI   `protobuf:"bytes,1,opt,name=vmi" json:"vmi,omitempty"`
	Options []byte `protobuf:"bytes,2,opt,name=options,proto3" json:"options,omitempty"`
}

func (m *MigrationRequest) Reset()                    { *m = MigrationRequest{} }
func (m *MigrationRequest) String() string            { return proto.CompactTextString(m) }
func (*MigrationRequest) ProtoMessage()               {}
func (*MigrationRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *MigrationRequest) GetVmi() *VMI {
	if m != nil {
		return m.Vmi
	}
	return nil
}

func (m *MigrationRequest) GetOptions() []byte {
	if m != nil {
		return m.Options
	}
	return nil
}

type EmptyRequest struct {
}

func (m *EmptyRequest) Reset()                    { *m = EmptyRequest{} }
func (m *EmptyRequest) String() string            { return proto.CompactTextString(m) }
func (*EmptyRequest) ProtoMessage()               {}
func (*EmptyRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type Response struct {
	Success bool   `protobuf:"varint,1,opt,name=success" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
}

func (m *Response) Reset()                    { *m = Response{} }
func (m *Response) String() string            { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()               {}
func (*Response) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *Response) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *Response) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type DomainResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Domain   string    `protobuf:"bytes,2,opt,name=domain" json:"domain,omitempty"`
}

func (m *DomainResponse) Reset()                    { *m = DomainResponse{} }
func (m *DomainResponse) String() string            { return proto.CompactTextString(m) }
func (*DomainResponse) ProtoMessage()               {}
func (*DomainResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *DomainResponse) GetResponse() *Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (m *DomainResponse) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	return ""
}

type DomainStatsResponse struct {
	Response    *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	DomainStats string    `protobuf:"bytes,2,opt,name=domainStats" json:"domainStats,omitempty"`
}

func (m *DomainStatsResponse) Reset()                    { *m = DomainStatsResponse{} }
func (m *DomainStatsResponse) String() string            { return proto.CompactTextString(m) }
func (*DomainStatsResponse) ProtoMessage()               {}
func (*DomainStatsResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *DomainStatsResponse) GetResponse() *Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (m *DomainStatsResponse) GetDomainStats() string {
	if m != nil {
		return m.DomainStats
	}
	return ""
}

func init() {
	proto.RegisterType((*VMI)(nil), "kubevirt.cmd.v1.VMI")
	proto.RegisterType((*SMBios)(nil), "kubevirt.cmd.v1.SMBios")
	proto.RegisterType((*HostProcessorsMap)(nil), "kubevirt.cmd.v1.HostProcessorsMap")
	proto.RegisterType((*VirtualMachineOptions)(nil), "kubevirt.cmd.v1.VirtualMachineOptions")
	proto.RegisterType((*VMIRequest)(nil), "kubevirt.cmd.v1.VMIRequest")
	proto.RegisterType((*MigrationRequest)(nil), "kubevirt.cmd.v1.MigrationRequest")
	proto.RegisterType((*EmptyRequest)(nil), "kubevirt.cmd.v1.EmptyRequest")
	proto.RegisterType((*Response)(nil), "kubevirt.cmd.v1.Response")
	proto.RegisterType((*DomainResponse)(nil), "kubevirt.cmd.v1.DomainResponse")
	proto.RegisterType((*DomainStatsResponse)(nil), "kubevirt.cmd.v1.DomainStatsResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Cmd service

type CmdClient interface {
	SyncVirtualMachine(ctx context.Context, in *VMIRequest, opts ...grpc.CallOption) (*Response, error)
	ShutdownVirtualMachine(ctx context.Context, in *VMIRequest, opts ...grpc.CallOption) (*Response, error)
	KillVirtualMachine(ctx context.Context, in *VMIRequest, opts ...grpc.CallOption) (*Response, error)
	DeleteVirtualMachine(ctx context.Context, in *VMIRequest, opts ...grpc.CallOption) (*Response, error)
	MigrateVirtualMachine(ctx context.Context, in *MigrationRequest, opts ...grpc.CallOption) (*Response, error)
	SyncMigrationTarget(ctx context.Context, in *VMIRequest, opts ...grpc.CallOption) (*Response, error)
	CancelVirtualMachineMigration(ctx context.Context, in *VMIRequest, opts ...grpc.CallOption) (*Response, error)
	GetDomain(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*DomainResponse, error)
	GetDomainStats(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*DomainStatsResponse, error)
	Ping(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*Response, error)
}

type cmdClient struct {
	cc *grpc.ClientConn
}

func NewCmdClient(cc *grpc.ClientConn) CmdClient {
	return &cmdClient{cc}
}

func (c *cmdClient) SyncVirtualMachine(ctx context.Context, in *VMIRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/kubevirt.cmd.v1.Cmd/SyncVirtualMachine", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cmdClient) ShutdownVirtualMachine(ctx context.Context, in *VMIRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/kubevirt.cmd.v1.Cmd/ShutdownVirtualMachine", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cmdClient) KillVirtualMachine(ctx context.Context, in *VMIRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/kubevirt.cmd.v1.Cmd/KillVirtualMachine", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cmdClient) DeleteVirtualMachine(ctx context.Context, in *VMIRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/kubevirt.cmd.v1.Cmd/DeleteVirtualMachine", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cmdClient) MigrateVirtualMachine(ctx context.Context, in *MigrationRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/kubevirt.cmd.v1.Cmd/MigrateVirtualMachine", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cmdClient) SyncMigrationTarget(ctx context.Context, in *VMIRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/kubevirt.cmd.v1.Cmd/SyncMigrationTarget", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cmdClient) CancelVirtualMachineMigration(ctx context.Context, in *VMIRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/kubevirt.cmd.v1.Cmd/CancelVirtualMachineMigration", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cmdClient) GetDomain(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*DomainResponse, error) {
	out := new(DomainResponse)
	err := grpc.Invoke(ctx, "/kubevirt.cmd.v1.Cmd/GetDomain", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cmdClient) GetDomainStats(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*DomainStatsResponse, error) {
	out := new(DomainStatsResponse)
	err := grpc.Invoke(ctx, "/kubevirt.cmd.v1.Cmd/GetDomainStats", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cmdClient) Ping(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/kubevirt.cmd.v1.Cmd/Ping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Cmd service

type CmdServer interface {
	SyncVirtualMachine(context.Context, *VMIRequest) (*Response, error)
	ShutdownVirtualMachine(context.Context, *VMIRequest) (*Response, error)
	KillVirtualMachine(context.Context, *VMIRequest) (*Response, error)
	DeleteVirtualMachine(context.Context, *VMIRequest) (*Response, error)
	MigrateVirtualMachine(context.Context, *MigrationRequest) (*Response, error)
	SyncMigrationTarget(context.Context, *VMIRequest) (*Response, error)
	CancelVirtualMachineMigration(context.Context, *VMIRequest) (*Response, error)
	GetDomain(context.Context, *EmptyRequest) (*DomainResponse, error)
	GetDomainStats(context.Context, *EmptyRequest) (*DomainStatsResponse, error)
	Ping(context.Context, *EmptyRequest) (*Response, error)
}

func RegisterCmdServer(s *grpc.Server, srv CmdServer) {
	s.RegisterService(&_Cmd_serviceDesc, srv)
}

func _Cmd_SyncVirtualMachine_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VMIRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CmdServer).SyncVirtualMachine(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubevirt.cmd.v1.Cmd/SyncVirtualMachine",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CmdServer).SyncVirtualMachine(ctx, req.(*VMIRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cmd_ShutdownVirtualMachine_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VMIRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CmdServer).ShutdownVirtualMachine(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubevirt.cmd.v1.Cmd/ShutdownVirtualMachine",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CmdServer).ShutdownVirtualMachine(ctx, req.(*VMIRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cmd_KillVirtualMachine_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VMIRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CmdServer).KillVirtualMachine(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubevirt.cmd.v1.Cmd/KillVirtualMachine",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CmdServer).KillVirtualMachine(ctx, req.(*VMIRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cmd_DeleteVirtualMachine_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VMIRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CmdServer).DeleteVirtualMachine(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubevirt.cmd.v1.Cmd/DeleteVirtualMachine",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CmdServer).DeleteVirtualMachine(ctx, req.(*VMIRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cmd_MigrateVirtualMachine_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MigrationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CmdServer).MigrateVirtualMachine(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubevirt.cmd.v1.Cmd/MigrateVirtualMachine",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CmdServer).MigrateVirtualMachine(ctx, req.(*MigrationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cmd_SyncMigrationTarget_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VMIRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CmdServer).SyncMigrationTarget(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubevirt.cmd.v1.Cmd/SyncMigrationTarget",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CmdServer).SyncMigrationTarget(ctx, req.(*VMIRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cmd_CancelVirtualMachineMigration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VMIRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CmdServer).CancelVirtualMachineMigration(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubevirt.cmd.v1.Cmd/CancelVirtualMachineMigration",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CmdServer).CancelVirtualMachineMigration(ctx, req.(*VMIRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cmd_GetDomain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CmdServer).GetDomain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubevirt.cmd.v1.Cmd/GetDomain",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CmdServer).GetDomain(ctx, req.(*EmptyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cmd_GetDomainStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CmdServer).GetDomainStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubevirt.cmd.v1.Cmd/GetDomainStats",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CmdServer).GetDomainStats(ctx, req.(*EmptyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cmd_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CmdServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubevirt.cmd.v1.Cmd/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CmdServer).Ping(ctx, req.(*EmptyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Cmd_serviceDesc = grpc.ServiceDesc{
	ServiceName: "kubevirt.cmd.v1.Cmd",
	HandlerType: (*CmdServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SyncVirtualMachine",
			Handler:    _Cmd_SyncVirtualMachine_Handler,
		},
		{
			MethodName: "ShutdownVirtualMachine",
			Handler:    _Cmd_ShutdownVirtualMachine_Handler,
		},
		{
			MethodName: "KillVirtualMachine",
			Handler:    _Cmd_KillVirtualMachine_Handler,
		},
		{
			MethodName: "DeleteVirtualMachine",
			Handler:    _Cmd_DeleteVirtualMachine_Handler,
		},
		{
			MethodName: "MigrateVirtualMachine",
			Handler:    _Cmd_MigrateVirtualMachine_Handler,
		},
		{
			MethodName: "SyncMigrationTarget",
			Handler:    _Cmd_SyncMigrationTarget_Handler,
		},
		{
			MethodName: "CancelVirtualMachineMigration",
			Handler:    _Cmd_CancelVirtualMachineMigration_Handler,
		},
		{
			MethodName: "GetDomain",
			Handler:    _Cmd_GetDomain_Handler,
		},
		{
			MethodName: "GetDomainStats",
			Handler:    _Cmd_GetDomainStats_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _Cmd_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/handler-launcher-com/cmd/v1/cmd.proto",
}

func init() { proto.RegisterFile("pkg/handler-launcher-com/cmd/v1/cmd.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 607 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x95, 0xdd, 0x4e, 0xd4, 0x40,
	0x14, 0xc7, 0x59, 0x16, 0x17, 0x38, 0xbb, 0x22, 0x0e, 0x1f, 0x56, 0x0c, 0x01, 0x27, 0x86, 0xe8,
	0x05, 0x4b, 0xc0, 0x70, 0x6b, 0x0c, 0x60, 0x04, 0x49, 0x81, 0x74, 0x09, 0x46, 0x6f, 0xcc, 0xd0,
	0x0e, 0xbb, 0x13, 0xda, 0x99, 0x3a, 0x33, 0xad, 0xe1, 0x15, 0x7c, 0x18, 0x9f, 0xc0, 0x87, 0x33,
	0x33, 0x6d, 0x97, 0xed, 0x96, 0x8f, 0x98, 0xdd, 0xab, 0xed, 0xf9, 0xfa, 0xfd, 0xcf, 0x39, 0x3d,
	0xcd, 0xc2, 0xbb, 0xf8, 0xba, 0xbb, 0xd5, 0x23, 0x3c, 0x08, 0xa9, 0xdc, 0x0c, 0x49, 0xc2, 0xfd,
	0x1e, 0x95, 0x9b, 0xbe, 0x88, 0xb6, 0xfc, 0x28, 0xd8, 0x4a, 0xb7, 0xcd, 0x4f, 0x3b, 0x96, 0x42,
	0x0b, 0xf4, 0xec, 0x3a, 0xb9, 0xa4, 0x29, 0x93, 0xba, 0x6d, 0x7c, 0xe9, 0x36, 0x5e, 0x83, 0xfa,
	0x85, 0x7b, 0x84, 0x1c, 0x98, 0x4e, 0x23, 0xf6, 0x45, 0x09, 0xee, 0xd4, 0xd6, 0x6b, 0x6f, 0x5b,
	0x5e, 0x61, 0xe2, 0xdf, 0x35, 0x68, 0x74, 0xdc, 0x3d, 0x26, 0x14, 0xc2, 0xd0, 0x8a, 0x08, 0x4f,
	0xae, 0x88, 0xaf, 0x13, 0x49, 0xa5, 0xcd, 0x9c, 0xf5, 0x4a, 0x3e, 0x03, 0x8a, 0xa5, 0x08, 0x12,
	0x5f, 0x3b, 0x93, 0x36, 0x5c, 0x98, 0x56, 0x82, 0x4a, 0xc5, 0x04, 0x77, 0xea, 0x59, 0x24, 0x37,
	0xd1, 0x3c, 0xd4, 0xd5, 0x75, 0xe2, 0x4c, 0x59, 0xaf, 0x79, 0x44, 0xcb, 0xd0, 0xb8, 0x22, 0x11,
	0x0b, 0x6f, 0x9c, 0x27, 0xd6, 0x99, 0x5b, 0x78, 0x17, 0x9e, 0x1f, 0x0a, 0xa5, 0xcf, 0xa4, 0xf0,
	0xa9, 0x52, 0x42, 0x2a, 0x97, 0xc4, 0x68, 0x1d, 0x9a, 0xb1, 0x14, 0xbe, 0x4b, 0xe2, 0x81, 0xfe,
	0x07, 0x5d, 0xf8, 0x4f, 0x0d, 0x96, 0x2e, 0x98, 0xd4, 0x09, 0x09, 0x5d, 0xe2, 0xf7, 0x18, 0xa7,
	0xa7, 0xb1, 0x66, 0x82, 0x2b, 0x74, 0x0c, 0x8b, 0xe5, 0x40, 0x36, 0xaa, 0x85, 0x34, 0x77, 0x5e,
	0xb4, 0x87, 0xd6, 0xd5, 0xce, 0xc2, 0xde, 0x9d, 0x45, 0xe8, 0x10, 0x9e, 0x96, 0x3a, 0xb3, 0x1b,
	0x68, 0xee, 0xe0, 0x0a, 0xa5, 0x32, 0x83, 0x57, 0x2e, 0xc4, 0x29, 0xc0, 0x85, 0x7b, 0xe4, 0xd1,
	0x9f, 0x09, 0x55, 0x1a, 0x6d, 0x40, 0x3d, 0x8d, 0x58, 0xde, 0xd3, 0x62, 0x85, 0x66, 0x32, 0x4d,
	0x02, 0xfa, 0x08, 0xd3, 0x22, 0x9b, 0x2b, 0x57, 0xde, 0xa8, 0xe6, 0xde, 0xb5, 0x05, 0xaf, 0x28,
	0xc3, 0xe7, 0x30, 0xef, 0xb2, 0xae, 0x24, 0xc6, 0xfa, 0x5f, 0x75, 0xa7, 0xac, 0xde, 0xba, 0xa5,
	0xce, 0x41, 0xeb, 0x53, 0x14, 0xeb, 0x9b, 0x9c, 0x88, 0x3f, 0xc0, 0x8c, 0x47, 0x55, 0x2c, 0xb8,
	0xa2, 0xa6, 0x4a, 0x25, 0xbe, 0x19, 0xdd, 0x2a, 0xcc, 0x78, 0x85, 0x69, 0x22, 0x11, 0x55, 0x8a,
	0x74, 0x69, 0x71, 0x49, 0xb9, 0x89, 0x7f, 0xc0, 0xdc, 0x81, 0x88, 0x08, 0xe3, 0x7d, 0xca, 0x2e,
	0xcc, 0xc8, 0xfc, 0x39, 0x6f, 0xf4, 0x65, 0xa5, 0xd1, 0x22, 0xd9, 0xeb, 0xa7, 0x9a, 0x33, 0x0b,
	0x2c, 0x28, 0x57, 0xc8, 0x2d, 0xcc, 0x61, 0x21, 0x13, 0xe8, 0x68, 0xa2, 0xd5, 0xa8, 0x2a, 0xeb,
	0xd0, 0x0c, 0x6e, 0x69, 0xb9, 0xd4, 0xa0, 0x6b, 0xe7, 0x6f, 0x03, 0xea, 0xfb, 0x51, 0x80, 0x4e,
	0x00, 0x75, 0x6e, 0xb8, 0x5f, 0x7e, 0x49, 0xe8, 0xd5, 0x9d, 0x3b, 0xcf, 0x76, 0xb9, 0x72, 0x7f,
	0x07, 0x78, 0x02, 0x79, 0xb0, 0xdc, 0xe9, 0x25, 0x3a, 0x10, 0xbf, 0xf8, 0xd8, 0x98, 0x27, 0x80,
	0x8e, 0x59, 0x18, 0x8e, 0x8d, 0x77, 0x06, 0x8b, 0x07, 0x34, 0xa4, 0x9a, 0x8e, 0x8d, 0xf8, 0x15,
	0x96, 0xb2, 0x23, 0x1e, 0x46, 0xbe, 0xae, 0x54, 0x0d, 0x1f, 0xfb, 0xc3, 0xe0, 0x53, 0x58, 0x30,
	0xaf, 0xa7, 0x5f, 0x74, 0x4e, 0x64, 0x97, 0xea, 0x11, 0x3a, 0xfd, 0x06, 0xab, 0xfb, 0x84, 0xfb,
	0x74, 0x68, 0x9b, 0x7d, 0x81, 0x11, 0xd0, 0x2e, 0xcc, 0x7e, 0xa6, 0x3a, 0xbb, 0x62, 0xb4, 0x5a,
	0xc9, 0x1c, 0xfc, 0x1e, 0x57, 0xd6, 0x2a, 0xe1, 0xf2, 0xe7, 0x65, 0x77, 0x3a, 0xd7, 0xc7, 0xd9,
	0x9b, 0x7d, 0x8c, 0xf9, 0xe6, 0x1e, 0x66, 0xe9, 0x8b, 0xc2, 0x13, 0x68, 0x0f, 0xa6, 0xce, 0x18,
	0xef, 0x3e, 0x86, 0x7b, 0x68, 0xd6, 0xbd, 0xa9, 0xef, 0x93, 0xe9, 0xf6, 0x65, 0xc3, 0xfe, 0xc3,
	0xbd, 0xff, 0x17, 0x00, 0x00, 0xff, 0xff, 0x44, 0x9f, 0x7d, 0x49, 0x0e, 0x07, 0x00, 0x00,
}
