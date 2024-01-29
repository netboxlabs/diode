// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        (unknown)
// source: diode/v1/device.proto

package diodepb

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// A device
type Device struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name       string     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	DeviceFqdn string     `protobuf:"bytes,2,opt,name=device_fqdn,json=deviceFqdn,proto3" json:"device_fqdn,omitempty"`
	DeviceType *anypb.Any `protobuf:"bytes,3,opt,name=device_type,json=deviceType,proto3" json:"device_type,omitempty"`
	Role       *anypb.Any `protobuf:"bytes,4,opt,name=role,proto3" json:"role,omitempty"`
	Platform   *anypb.Any `protobuf:"bytes,5,opt,name=platform,proto3" json:"platform,omitempty"`
	Serial     string     `protobuf:"bytes,6,opt,name=serial,proto3" json:"serial,omitempty"`
	Site       *anypb.Any `protobuf:"bytes,7,opt,name=site,proto3" json:"site,omitempty"`
	VcPosition int32      `protobuf:"varint,8,opt,name=vc_position,json=vcPosition,proto3" json:"vc_position,omitempty"`
}

func (x *Device) Reset() {
	*x = Device{}
	if protoimpl.UnsafeEnabled {
		mi := &file_diode_v1_device_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Device) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Device) ProtoMessage() {}

func (x *Device) ProtoReflect() protoreflect.Message {
	mi := &file_diode_v1_device_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Device.ProtoReflect.Descriptor instead.
func (*Device) Descriptor() ([]byte, []int) {
	return file_diode_v1_device_proto_rawDescGZIP(), []int{0}
}

func (x *Device) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Device) GetDeviceFqdn() string {
	if x != nil {
		return x.DeviceFqdn
	}
	return ""
}

func (x *Device) GetDeviceType() *anypb.Any {
	if x != nil {
		return x.DeviceType
	}
	return nil
}

func (x *Device) GetRole() *anypb.Any {
	if x != nil {
		return x.Role
	}
	return nil
}

func (x *Device) GetPlatform() *anypb.Any {
	if x != nil {
		return x.Platform
	}
	return nil
}

func (x *Device) GetSerial() string {
	if x != nil {
		return x.Serial
	}
	return ""
}

func (x *Device) GetSite() *anypb.Any {
	if x != nil {
		return x.Site
	}
	return nil
}

func (x *Device) GetVcPosition() int32 {
	if x != nil {
		return x.VcPosition
	}
	return 0
}

var File_diode_v1_device_proto protoreflect.FileDescriptor

var file_diode_v1_device_proto_rawDesc = []byte{
	0x0a, 0x15, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2e, 0x76,
	0x31, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61,
	0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xfb, 0x02, 0x0a, 0x06, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x1b, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x07,
	0xfa, 0x42, 0x04, 0x72, 0x02, 0x18, 0x40, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x2b, 0x0a,
	0x0b, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x66, 0x71, 0x64, 0x6e, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x0a, 0xfa, 0x42, 0x07, 0x72, 0x05, 0x10, 0x01, 0x18, 0xff, 0x01, 0x52, 0x0a,
	0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x46, 0x71, 0x64, 0x6e, 0x12, 0x3f, 0x0a, 0x0b, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x41, 0x6e, 0x79, 0x42, 0x08, 0xfa, 0x42, 0x05, 0xa2, 0x01, 0x02, 0x08, 0x01, 0x52,
	0x0a, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x32, 0x0a, 0x04, 0x72,
	0x6f, 0x6c, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x42,
	0x08, 0xfa, 0x42, 0x05, 0xa2, 0x01, 0x02, 0x08, 0x01, 0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x12,
	0x30, 0x0a, 0x08, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x08, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72,
	0x6d, 0x12, 0x1f, 0x0a, 0x06, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x72, 0x02, 0x18, 0x32, 0x52, 0x06, 0x73, 0x65, 0x72, 0x69,
	0x61, 0x6c, 0x12, 0x32, 0x0a, 0x04, 0x73, 0x69, 0x74, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x42, 0x08, 0xfa, 0x42, 0x05, 0xa2, 0x01, 0x02, 0x08, 0x01,
	0x52, 0x04, 0x73, 0x69, 0x74, 0x65, 0x12, 0x2b, 0x0a, 0x0b, 0x76, 0x63, 0x5f, 0x70, 0x6f, 0x73,
	0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x42, 0x0a, 0xfa, 0x42, 0x07,
	0x1a, 0x05, 0x18, 0xff, 0x01, 0x28, 0x00, 0x52, 0x0a, 0x76, 0x63, 0x50, 0x6f, 0x73, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x42, 0x44, 0x5a, 0x42, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6e, 0x65, 0x74, 0x62, 0x6f, 0x78, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x64, 0x69, 0x6f,
	0x64, 0x65, 0x2d, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x64, 0x69, 0x6f, 0x64,
	0x65, 0x2d, 0x73, 0x64, 0x6b, 0x2d, 0x67, 0x6f, 0x2f, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2f, 0x76,
	0x31, 0x2f, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_diode_v1_device_proto_rawDescOnce sync.Once
	file_diode_v1_device_proto_rawDescData = file_diode_v1_device_proto_rawDesc
)

func file_diode_v1_device_proto_rawDescGZIP() []byte {
	file_diode_v1_device_proto_rawDescOnce.Do(func() {
		file_diode_v1_device_proto_rawDescData = protoimpl.X.CompressGZIP(file_diode_v1_device_proto_rawDescData)
	})
	return file_diode_v1_device_proto_rawDescData
}

var file_diode_v1_device_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_diode_v1_device_proto_goTypes = []interface{}{
	(*Device)(nil),    // 0: diode.v1.Device
	(*anypb.Any)(nil), // 1: google.protobuf.Any
}
var file_diode_v1_device_proto_depIdxs = []int32{
	1, // 0: diode.v1.Device.device_type:type_name -> google.protobuf.Any
	1, // 1: diode.v1.Device.role:type_name -> google.protobuf.Any
	1, // 2: diode.v1.Device.platform:type_name -> google.protobuf.Any
	1, // 3: diode.v1.Device.site:type_name -> google.protobuf.Any
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_diode_v1_device_proto_init() }
func file_diode_v1_device_proto_init() {
	if File_diode_v1_device_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_diode_v1_device_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Device); i {
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
			RawDescriptor: file_diode_v1_device_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_diode_v1_device_proto_goTypes,
		DependencyIndexes: file_diode_v1_device_proto_depIdxs,
		MessageInfos:      file_diode_v1_device_proto_msgTypes,
	}.Build()
	File_diode_v1_device_proto = out.File
	file_diode_v1_device_proto_rawDesc = nil
	file_diode_v1_device_proto_goTypes = nil
	file_diode_v1_device_proto_depIdxs = nil
}