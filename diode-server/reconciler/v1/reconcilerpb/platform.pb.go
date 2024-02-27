// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        (unknown)
// source: reconciler/v1/platform.proto

package reconcilerpb

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

// A platform
type Platform struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name    string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Slug    string `protobuf:"bytes,3,opt,name=slug,proto3" json:"slug,omitempty"`
	Display string `protobuf:"bytes,4,opt,name=display,proto3" json:"display,omitempty"`
	Url     string `protobuf:"bytes,5,opt,name=url,proto3" json:"url,omitempty"`
}

func (x *Platform) Reset() {
	*x = Platform{}
	if protoimpl.UnsafeEnabled {
		mi := &file_reconciler_v1_platform_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Platform) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Platform) ProtoMessage() {}

func (x *Platform) ProtoReflect() protoreflect.Message {
	mi := &file_reconciler_v1_platform_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Platform.ProtoReflect.Descriptor instead.
func (*Platform) Descriptor() ([]byte, []int) {
	return file_reconciler_v1_platform_proto_rawDescGZIP(), []int{0}
}

func (x *Platform) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Platform) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Platform) GetSlug() string {
	if x != nil {
		return x.Slug
	}
	return ""
}

func (x *Platform) GetDisplay() string {
	if x != nil {
		return x.Display
	}
	return ""
}

func (x *Platform) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

var File_reconciler_v1_platform_proto protoreflect.FileDescriptor

var file_reconciler_v1_platform_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x72, 0x65, 0x63, 0x6f, 0x6e, 0x63, 0x69, 0x6c, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f,
	0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d,
	0x72, 0x65, 0x63, 0x6f, 0x6e, 0x63, 0x69, 0x6c, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x1a, 0x17, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa9, 0x01, 0x0a, 0x08, 0x50, 0x6c, 0x61, 0x74, 0x66,
	0x6f, 0x72, 0x6d, 0x12, 0x17, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x42,
	0x07, 0xfa, 0x42, 0x04, 0x32, 0x02, 0x28, 0x01, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1d, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x09, 0xfa, 0x42, 0x06, 0x72,
	0x04, 0x10, 0x01, 0x18, 0x64, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x2f, 0x0a, 0x04, 0x73,
	0x6c, 0x75, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x1b, 0xfa, 0x42, 0x18, 0x72, 0x16,
	0x10, 0x01, 0x18, 0x64, 0x32, 0x10, 0x5e, 0x5b, 0x2d, 0x61, 0x2d, 0x7a, 0x41, 0x2d, 0x5a, 0x30,
	0x2d, 0x39, 0x5f, 0x5d, 0x2b, 0x24, 0x52, 0x04, 0x73, 0x6c, 0x75, 0x67, 0x12, 0x18, 0x0a, 0x07,
	0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x64,
	0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x12, 0x1a, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x72, 0x03, 0x88, 0x01, 0x01, 0x52, 0x03, 0x75,
	0x72, 0x6c, 0x42, 0x45, 0x5a, 0x43, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x6e, 0x65, 0x74, 0x62, 0x6f, 0x78, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x64, 0x69, 0x6f, 0x64,
	0x65, 0x2f, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x72,
	0x65, 0x63, 0x6f, 0x6e, 0x63, 0x69, 0x6c, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x63,
	0x6f, 0x6e, 0x63, 0x69, 0x6c, 0x65, 0x72, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_reconciler_v1_platform_proto_rawDescOnce sync.Once
	file_reconciler_v1_platform_proto_rawDescData = file_reconciler_v1_platform_proto_rawDesc
)

func file_reconciler_v1_platform_proto_rawDescGZIP() []byte {
	file_reconciler_v1_platform_proto_rawDescOnce.Do(func() {
		file_reconciler_v1_platform_proto_rawDescData = protoimpl.X.CompressGZIP(file_reconciler_v1_platform_proto_rawDescData)
	})
	return file_reconciler_v1_platform_proto_rawDescData
}

var file_reconciler_v1_platform_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_reconciler_v1_platform_proto_goTypes = []interface{}{
	(*Platform)(nil), // 0: reconciler.v1.Platform
}
var file_reconciler_v1_platform_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_reconciler_v1_platform_proto_init() }
func file_reconciler_v1_platform_proto_init() {
	if File_reconciler_v1_platform_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_reconciler_v1_platform_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Platform); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_reconciler_v1_platform_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_reconciler_v1_platform_proto_goTypes,
		DependencyIndexes: file_reconciler_v1_platform_proto_depIdxs,
		MessageInfos:      file_reconciler_v1_platform_proto_msgTypes,
	}.Build()
	File_reconciler_v1_platform_proto = out.File
	file_reconciler_v1_platform_proto_rawDesc = nil
	file_reconciler_v1_platform_proto_goTypes = nil
	file_reconciler_v1_platform_proto_depIdxs = nil
}