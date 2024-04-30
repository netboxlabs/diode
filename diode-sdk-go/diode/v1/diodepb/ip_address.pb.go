// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        (unknown)
// source: diode/v1/ip_address.proto

package diodepb

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// An IP address.
type IPAddress struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// Types that are assignable to AssignedObject:
	//
	//	*IPAddress_Interface
	AssignedObject isIPAddress_AssignedObject `protobuf_oneof:"assigned_object"`
	Status         string                     `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	Role           string                     `protobuf:"bytes,4,opt,name=role,proto3" json:"role,omitempty"`
	DnsName        string                     `protobuf:"bytes,5,opt,name=dns_name,json=dnsName,proto3" json:"dns_name,omitempty"`
	Description    string                     `protobuf:"bytes,6,opt,name=description,proto3" json:"description,omitempty"`
	Comments       string                     `protobuf:"bytes,7,opt,name=comments,proto3" json:"comments,omitempty"`
	Tags           []*Tag                     `protobuf:"bytes,8,rep,name=tags,proto3" json:"tags,omitempty"`
}

func (x *IPAddress) Reset() {
	*x = IPAddress{}
	if protoimpl.UnsafeEnabled {
		mi := &file_diode_v1_ip_address_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IPAddress) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IPAddress) ProtoMessage() {}

func (x *IPAddress) ProtoReflect() protoreflect.Message {
	mi := &file_diode_v1_ip_address_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IPAddress.ProtoReflect.Descriptor instead.
func (*IPAddress) Descriptor() ([]byte, []int) {
	return file_diode_v1_ip_address_proto_rawDescGZIP(), []int{0}
}

func (x *IPAddress) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (m *IPAddress) GetAssignedObject() isIPAddress_AssignedObject {
	if m != nil {
		return m.AssignedObject
	}
	return nil
}

func (x *IPAddress) GetInterface() *Interface {
	if x, ok := x.GetAssignedObject().(*IPAddress_Interface); ok {
		return x.Interface
	}
	return nil
}

func (x *IPAddress) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *IPAddress) GetRole() string {
	if x != nil {
		return x.Role
	}
	return ""
}

func (x *IPAddress) GetDnsName() string {
	if x != nil {
		return x.DnsName
	}
	return ""
}

func (x *IPAddress) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *IPAddress) GetComments() string {
	if x != nil {
		return x.Comments
	}
	return ""
}

func (x *IPAddress) GetTags() []*Tag {
	if x != nil {
		return x.Tags
	}
	return nil
}

type isIPAddress_AssignedObject interface {
	isIPAddress_AssignedObject()
}

type IPAddress_Interface struct {
	Interface *Interface `protobuf:"bytes,2,opt,name=interface,proto3,oneof"`
}

func (*IPAddress_Interface) isIPAddress_AssignedObject() {}

var File_diode_v1_ip_address_proto protoreflect.FileDescriptor

var file_diode_v1_ip_address_proto_rawDesc = []byte{
	0x0a, 0x19, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x69, 0x70, 0x5f, 0x61, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x64, 0x69, 0x6f,
	0x64, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x18, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2f, 0x76, 0x31, 0x2f,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x12, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x61, 0x67, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61,
	0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd3, 0x03, 0x0a,
	0x09, 0x49, 0x50, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x21, 0x0a, 0x07, 0x61, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x07, 0xfa, 0x42, 0x04,
	0x72, 0x02, 0x70, 0x01, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x33, 0x0a,
	0x09, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x13, 0x2e, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x74, 0x65,
	0x72, 0x66, 0x61, 0x63, 0x65, 0x48, 0x00, 0x52, 0x09, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61,
	0x63, 0x65, 0x12, 0x48, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x30, 0xfa, 0x42, 0x2d, 0x72, 0x2b, 0x52, 0x06, 0x61, 0x63, 0x74, 0x69, 0x76,
	0x65, 0x52, 0x08, 0x72, 0x65, 0x73, 0x65, 0x72, 0x76, 0x65, 0x64, 0x52, 0x0a, 0x64, 0x65, 0x70,
	0x72, 0x65, 0x63, 0x61, 0x74, 0x65, 0x64, 0x52, 0x04, 0x64, 0x68, 0x63, 0x70, 0x52, 0x05, 0x73,
	0x6c, 0x61, 0x61, 0x63, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x54, 0x0a, 0x04,
	0x72, 0x6f, 0x6c, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x42, 0x40, 0xfa, 0x42, 0x3d, 0x72,
	0x3b, 0x52, 0x08, 0x6c, 0x6f, 0x6f, 0x70, 0x62, 0x61, 0x63, 0x6b, 0x52, 0x09, 0x73, 0x65, 0x63,
	0x6f, 0x6e, 0x64, 0x61, 0x72, 0x79, 0x52, 0x07, 0x61, 0x6e, 0x79, 0x63, 0x61, 0x73, 0x74, 0x52,
	0x03, 0x76, 0x69, 0x70, 0x52, 0x04, 0x76, 0x72, 0x72, 0x70, 0x52, 0x04, 0x68, 0x73, 0x72, 0x70,
	0x52, 0x04, 0x67, 0x6c, 0x62, 0x70, 0x52, 0x04, 0x63, 0x61, 0x72, 0x70, 0x52, 0x04, 0x72, 0x6f,
	0x6c, 0x65, 0x12, 0x50, 0x0a, 0x08, 0x64, 0x6e, 0x73, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x35, 0xfa, 0x42, 0x32, 0x72, 0x30, 0x18, 0xff, 0x01, 0x32, 0x2b,
	0x5e, 0x28, 0x5b, 0x30, 0x2d, 0x39, 0x41, 0x2d, 0x5a, 0x61, 0x2d, 0x7a, 0x5f, 0x2d, 0x5d, 0x2b,
	0x7c, 0x5c, 0x2a, 0x29, 0x28, 0x5c, 0x2e, 0x5b, 0x30, 0x2d, 0x39, 0x41, 0x2d, 0x5a, 0x61, 0x2d,
	0x7a, 0x5f, 0x2d, 0x5d, 0x2b, 0x29, 0x2a, 0x5c, 0x2e, 0x3f, 0x24, 0x52, 0x07, 0x64, 0x6e, 0x73,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x2a, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x72, 0x03,
	0x18, 0xc8, 0x01, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x21, 0x0a, 0x04,
	0x74, 0x61, 0x67, 0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x64, 0x69, 0x6f,
	0x64, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x61, 0x67, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x42,
	0x11, 0x0a, 0x0f, 0x61, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x5f, 0x6f, 0x62, 0x6a, 0x65,
	0x63, 0x74, 0x42, 0x3b, 0x5a, 0x39, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x6e, 0x65, 0x74, 0x62, 0x6f, 0x78, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x64, 0x69, 0x6f, 0x64,
	0x65, 0x2f, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2d, 0x73, 0x64, 0x6b, 0x2d, 0x67, 0x6f, 0x2f, 0x64,
	0x69, 0x6f, 0x64, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_diode_v1_ip_address_proto_rawDescOnce sync.Once
	file_diode_v1_ip_address_proto_rawDescData = file_diode_v1_ip_address_proto_rawDesc
)

func file_diode_v1_ip_address_proto_rawDescGZIP() []byte {
	file_diode_v1_ip_address_proto_rawDescOnce.Do(func() {
		file_diode_v1_ip_address_proto_rawDescData = protoimpl.X.CompressGZIP(file_diode_v1_ip_address_proto_rawDescData)
	})
	return file_diode_v1_ip_address_proto_rawDescData
}

var file_diode_v1_ip_address_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_diode_v1_ip_address_proto_goTypes = []interface{}{
	(*IPAddress)(nil), // 0: diode.v1.IPAddress
	(*Interface)(nil), // 1: diode.v1.Interface
	(*Tag)(nil),       // 2: diode.v1.Tag
}
var file_diode_v1_ip_address_proto_depIdxs = []int32{
	1, // 0: diode.v1.IPAddress.interface:type_name -> diode.v1.Interface
	2, // 1: diode.v1.IPAddress.tags:type_name -> diode.v1.Tag
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_diode_v1_ip_address_proto_init() }
func file_diode_v1_ip_address_proto_init() {
	if File_diode_v1_ip_address_proto != nil {
		return
	}
	file_diode_v1_interface_proto_init()
	file_diode_v1_tag_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_diode_v1_ip_address_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IPAddress); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_diode_v1_ip_address_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*IPAddress_Interface)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_diode_v1_ip_address_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_diode_v1_ip_address_proto_goTypes,
		DependencyIndexes: file_diode_v1_ip_address_proto_depIdxs,
		MessageInfos:      file_diode_v1_ip_address_proto_msgTypes,
	}.Build()
	File_diode_v1_ip_address_proto = out.File
	file_diode_v1_ip_address_proto_rawDesc = nil
	file_diode_v1_ip_address_proto_goTypes = nil
	file_diode_v1_ip_address_proto_depIdxs = nil
}