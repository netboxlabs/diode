// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        (unknown)
// source: diode/v1/ingester.proto

package diodepb

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// An ingest entity wrapper
type Entity struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Data:
	//
	//	*Entity_Site
	//	*Entity_Platform
	//	*Entity_Manufacturer
	//	*Entity_Device
	//	*Entity_DeviceRole
	//	*Entity_DeviceType
	//	*Entity_Interface
	Data isEntity_Data `protobuf_oneof:"data"`
	// The timestamp of the data discovery at source
	Timestamp *timestamppb.Timestamp `protobuf:"bytes,8,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *Entity) Reset() {
	*x = Entity{}
	if protoimpl.UnsafeEnabled {
		mi := &file_diode_v1_ingester_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Entity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Entity) ProtoMessage() {}

func (x *Entity) ProtoReflect() protoreflect.Message {
	mi := &file_diode_v1_ingester_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Entity.ProtoReflect.Descriptor instead.
func (*Entity) Descriptor() ([]byte, []int) {
	return file_diode_v1_ingester_proto_rawDescGZIP(), []int{0}
}

func (m *Entity) GetData() isEntity_Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (x *Entity) GetSite() *Site {
	if x, ok := x.GetData().(*Entity_Site); ok {
		return x.Site
	}
	return nil
}

func (x *Entity) GetPlatform() *Platform {
	if x, ok := x.GetData().(*Entity_Platform); ok {
		return x.Platform
	}
	return nil
}

func (x *Entity) GetManufacturer() *Manufacturer {
	if x, ok := x.GetData().(*Entity_Manufacturer); ok {
		return x.Manufacturer
	}
	return nil
}

func (x *Entity) GetDevice() *Device {
	if x, ok := x.GetData().(*Entity_Device); ok {
		return x.Device
	}
	return nil
}

func (x *Entity) GetDeviceRole() *Role {
	if x, ok := x.GetData().(*Entity_DeviceRole); ok {
		return x.DeviceRole
	}
	return nil
}

func (x *Entity) GetDeviceType() *DeviceType {
	if x, ok := x.GetData().(*Entity_DeviceType); ok {
		return x.DeviceType
	}
	return nil
}

func (x *Entity) GetInterface() *Interface {
	if x, ok := x.GetData().(*Entity_Interface); ok {
		return x.Interface
	}
	return nil
}

func (x *Entity) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

type isEntity_Data interface {
	isEntity_Data()
}

type Entity_Site struct {
	Site *Site `protobuf:"bytes,1,opt,name=site,proto3,oneof"`
}

type Entity_Platform struct {
	Platform *Platform `protobuf:"bytes,2,opt,name=platform,proto3,oneof"`
}

type Entity_Manufacturer struct {
	Manufacturer *Manufacturer `protobuf:"bytes,3,opt,name=manufacturer,proto3,oneof"`
}

type Entity_Device struct {
	Device *Device `protobuf:"bytes,4,opt,name=device,proto3,oneof"`
}

type Entity_DeviceRole struct {
	DeviceRole *Role `protobuf:"bytes,5,opt,name=device_role,json=deviceRole,proto3,oneof"`
}

type Entity_DeviceType struct {
	DeviceType *DeviceType `protobuf:"bytes,6,opt,name=device_type,json=deviceType,proto3,oneof"`
}

type Entity_Interface struct {
	Interface *Interface `protobuf:"bytes,7,opt,name=interface,proto3,oneof"`
}

func (*Entity_Site) isEntity_Data() {}

func (*Entity_Platform) isEntity_Data() {}

func (*Entity_Manufacturer) isEntity_Data() {}

func (*Entity_Device) isEntity_Data() {}

func (*Entity_DeviceRole) isEntity_Data() {}

func (*Entity_DeviceType) isEntity_Data() {}

func (*Entity_Interface) isEntity_Data() {}

// The request to ingest the data
type IngestRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Stream             string    `protobuf:"bytes,1,opt,name=stream,proto3" json:"stream,omitempty"`
	Entities           []*Entity `protobuf:"bytes,2,rep,name=entities,proto3" json:"entities,omitempty"`
	Id                 string    `protobuf:"bytes,3,opt,name=id,proto3" json:"id,omitempty"`
	ProducerAppName    string    `protobuf:"bytes,4,opt,name=producer_app_name,json=producerAppName,proto3" json:"producer_app_name,omitempty"`
	ProducerAppVersion string    `protobuf:"bytes,5,opt,name=producer_app_version,json=producerAppVersion,proto3" json:"producer_app_version,omitempty"`
	SdkName            string    `protobuf:"bytes,6,opt,name=sdk_name,json=sdkName,proto3" json:"sdk_name,omitempty"`
	SdkVersion         string    `protobuf:"bytes,7,opt,name=sdk_version,json=sdkVersion,proto3" json:"sdk_version,omitempty"`
}

func (x *IngestRequest) Reset() {
	*x = IngestRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_diode_v1_ingester_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IngestRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IngestRequest) ProtoMessage() {}

func (x *IngestRequest) ProtoReflect() protoreflect.Message {
	mi := &file_diode_v1_ingester_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IngestRequest.ProtoReflect.Descriptor instead.
func (*IngestRequest) Descriptor() ([]byte, []int) {
	return file_diode_v1_ingester_proto_rawDescGZIP(), []int{1}
}

func (x *IngestRequest) GetStream() string {
	if x != nil {
		return x.Stream
	}
	return ""
}

func (x *IngestRequest) GetEntities() []*Entity {
	if x != nil {
		return x.Entities
	}
	return nil
}

func (x *IngestRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *IngestRequest) GetProducerAppName() string {
	if x != nil {
		return x.ProducerAppName
	}
	return ""
}

func (x *IngestRequest) GetProducerAppVersion() string {
	if x != nil {
		return x.ProducerAppVersion
	}
	return ""
}

func (x *IngestRequest) GetSdkName() string {
	if x != nil {
		return x.SdkName
	}
	return ""
}

func (x *IngestRequest) GetSdkVersion() string {
	if x != nil {
		return x.SdkVersion
	}
	return ""
}

// The response from the ingest request
type IngestResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Errors []string `protobuf:"bytes,1,rep,name=errors,proto3" json:"errors,omitempty"`
}

func (x *IngestResponse) Reset() {
	*x = IngestResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_diode_v1_ingester_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IngestResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IngestResponse) ProtoMessage() {}

func (x *IngestResponse) ProtoReflect() protoreflect.Message {
	mi := &file_diode_v1_ingester_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IngestResponse.ProtoReflect.Descriptor instead.
func (*IngestResponse) Descriptor() ([]byte, []int) {
	return file_diode_v1_ingester_proto_rawDescGZIP(), []int{2}
}

func (x *IngestResponse) GetErrors() []string {
	if x != nil {
		return x.Errors
	}
	return nil
}

var File_diode_v1_ingester_proto protoreflect.FileDescriptor

var file_diode_v1_ingester_proto_rawDesc = []byte{
	0x0a, 0x17, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x69, 0x6e, 0x67, 0x65, 0x73,
	0x74, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x64, 0x69, 0x6f, 0x64, 0x65,
	0x2e, 0x76, 0x31, 0x1a, 0x15, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1a, 0x64, 0x69, 0x6f, 0x64,
	0x65, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x18, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2f, 0x76, 0x31,
	0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1b, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x61, 0x6e, 0x75, 0x66,
	0x61, 0x63, 0x74, 0x75, 0x72, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x64,
	0x69, 0x6f, 0x64, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x13, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2f, 0x76, 0x31,
	0x2f, 0x72, 0x6f, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x13, 0x64, 0x69, 0x6f,
	0x64, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x69, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb9, 0x03, 0x0a, 0x06, 0x45,
	0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x24, 0x0a, 0x04, 0x73, 0x69, 0x74, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53,
	0x69, 0x74, 0x65, 0x48, 0x00, 0x52, 0x04, 0x73, 0x69, 0x74, 0x65, 0x12, 0x30, 0x0a, 0x08, 0x70,
	0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e,
	0x64, 0x69, 0x6f, 0x64, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72,
	0x6d, 0x48, 0x00, 0x52, 0x08, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x12, 0x3c, 0x0a,
	0x0c, 0x6d, 0x61, 0x6e, 0x75, 0x66, 0x61, 0x63, 0x74, 0x75, 0x72, 0x65, 0x72, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4d,
	0x61, 0x6e, 0x75, 0x66, 0x61, 0x63, 0x74, 0x75, 0x72, 0x65, 0x72, 0x48, 0x00, 0x52, 0x0c, 0x6d,
	0x61, 0x6e, 0x75, 0x66, 0x61, 0x63, 0x74, 0x75, 0x72, 0x65, 0x72, 0x12, 0x2a, 0x0a, 0x06, 0x64,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x64, 0x69,
	0x6f, 0x64, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x48, 0x00, 0x52,
	0x06, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x12, 0x31, 0x0a, 0x0b, 0x64, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x5f, 0x72, 0x6f, 0x6c, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x64,
	0x69, 0x6f, 0x64, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x6f, 0x6c, 0x65, 0x48, 0x00, 0x52, 0x0a,
	0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x52, 0x6f, 0x6c, 0x65, 0x12, 0x37, 0x0a, 0x0b, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x54, 0x79, 0x70, 0x65, 0x48, 0x00, 0x52, 0x0a, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x33, 0x0a, 0x09, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2e, 0x76,
	0x31, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x48, 0x00, 0x52, 0x09, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x12, 0x44, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x0a, 0xfa, 0x42, 0x07, 0xb2, 0x01, 0x04, 0x08,
	0x01, 0x38, 0x01, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x06,
	0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0xe4, 0x02, 0x0a, 0x0d, 0x49, 0x6e, 0x67, 0x65, 0x73,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x22, 0x0a, 0x06, 0x73, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x0a, 0xfa, 0x42, 0x07, 0x72, 0x05, 0x10,
	0x01, 0x18, 0xff, 0x01, 0x52, 0x06, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x39, 0x0a, 0x08,
	0x65, 0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10,
	0x2e, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x42, 0x0b, 0xfa, 0x42, 0x08, 0x92, 0x01, 0x05, 0x08, 0x01, 0x10, 0xe8, 0x07, 0x52, 0x08, 0x65,
	0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x12, 0x18, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x72, 0x03, 0xb0, 0x01, 0x01, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x36, 0x0a, 0x11, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x72, 0x5f, 0x61, 0x70,
	0x70, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x42, 0x0a, 0xfa, 0x42,
	0x07, 0x72, 0x05, 0x10, 0x01, 0x18, 0xff, 0x01, 0x52, 0x0f, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63,
	0x65, 0x72, 0x41, 0x70, 0x70, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x3c, 0x0a, 0x14, 0x70, 0x72, 0x6f,
	0x64, 0x75, 0x63, 0x65, 0x72, 0x5f, 0x61, 0x70, 0x70, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x42, 0x0a, 0xfa, 0x42, 0x07, 0x72, 0x05, 0x10, 0x01,
	0x18, 0xff, 0x01, 0x52, 0x12, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x72, 0x41, 0x70, 0x70,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x08, 0x73, 0x64, 0x6b, 0x5f, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x42, 0x0a, 0xfa, 0x42, 0x07, 0x72, 0x05,
	0x10, 0x01, 0x18, 0xff, 0x01, 0x52, 0x07, 0x73, 0x64, 0x6b, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x3d,
	0x0a, 0x0b, 0x73, 0x64, 0x6b, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x1c, 0xfa, 0x42, 0x19, 0x72, 0x17, 0x32, 0x15, 0x5e, 0x28, 0x5c, 0x64,
	0x29, 0x2b, 0x5c, 0x2e, 0x28, 0x5c, 0x64, 0x29, 0x2b, 0x5c, 0x2e, 0x28, 0x5c, 0x64, 0x29, 0x2b,
	0x24, 0x52, 0x0a, 0x73, 0x64, 0x6b, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x28, 0x0a,
	0x0e, 0x49, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x06, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x32, 0x50, 0x0a, 0x0f, 0x49, 0x6e, 0x67, 0x65, 0x73,
	0x74, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3d, 0x0a, 0x06, 0x49, 0x6e,
	0x67, 0x65, 0x73, 0x74, 0x12, 0x17, 0x2e, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2e, 0x76, 0x31, 0x2e,
	0x49, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e,
	0x64, 0x69, 0x6f, 0x64, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x67, 0x65, 0x73, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x3b, 0x5a, 0x39, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6e, 0x65, 0x74, 0x62, 0x6f, 0x78, 0x6c, 0x61,
	0x62, 0x73, 0x2f, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2f, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2d, 0x73,
	0x64, 0x6b, 0x2d, 0x67, 0x6f, 0x2f, 0x64, 0x69, 0x6f, 0x64, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x64,
	0x69, 0x6f, 0x64, 0x65, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_diode_v1_ingester_proto_rawDescOnce sync.Once
	file_diode_v1_ingester_proto_rawDescData = file_diode_v1_ingester_proto_rawDesc
)

func file_diode_v1_ingester_proto_rawDescGZIP() []byte {
	file_diode_v1_ingester_proto_rawDescOnce.Do(func() {
		file_diode_v1_ingester_proto_rawDescData = protoimpl.X.CompressGZIP(file_diode_v1_ingester_proto_rawDescData)
	})
	return file_diode_v1_ingester_proto_rawDescData
}

var file_diode_v1_ingester_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_diode_v1_ingester_proto_goTypes = []interface{}{
	(*Entity)(nil),                // 0: diode.v1.Entity
	(*IngestRequest)(nil),         // 1: diode.v1.IngestRequest
	(*IngestResponse)(nil),        // 2: diode.v1.IngestResponse
	(*Site)(nil),                  // 3: diode.v1.Site
	(*Platform)(nil),              // 4: diode.v1.Platform
	(*Manufacturer)(nil),          // 5: diode.v1.Manufacturer
	(*Device)(nil),                // 6: diode.v1.Device
	(*Role)(nil),                  // 7: diode.v1.Role
	(*DeviceType)(nil),            // 8: diode.v1.DeviceType
	(*Interface)(nil),             // 9: diode.v1.Interface
	(*timestamppb.Timestamp)(nil), // 10: google.protobuf.Timestamp
}
var file_diode_v1_ingester_proto_depIdxs = []int32{
	3,  // 0: diode.v1.Entity.site:type_name -> diode.v1.Site
	4,  // 1: diode.v1.Entity.platform:type_name -> diode.v1.Platform
	5,  // 2: diode.v1.Entity.manufacturer:type_name -> diode.v1.Manufacturer
	6,  // 3: diode.v1.Entity.device:type_name -> diode.v1.Device
	7,  // 4: diode.v1.Entity.device_role:type_name -> diode.v1.Role
	8,  // 5: diode.v1.Entity.device_type:type_name -> diode.v1.DeviceType
	9,  // 6: diode.v1.Entity.interface:type_name -> diode.v1.Interface
	10, // 7: diode.v1.Entity.timestamp:type_name -> google.protobuf.Timestamp
	0,  // 8: diode.v1.IngestRequest.entities:type_name -> diode.v1.Entity
	1,  // 9: diode.v1.IngesterService.Ingest:input_type -> diode.v1.IngestRequest
	2,  // 10: diode.v1.IngesterService.Ingest:output_type -> diode.v1.IngestResponse
	10, // [10:11] is the sub-list for method output_type
	9,  // [9:10] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_diode_v1_ingester_proto_init() }
func file_diode_v1_ingester_proto_init() {
	if File_diode_v1_ingester_proto != nil {
		return
	}
	file_diode_v1_device_proto_init()
	file_diode_v1_device_type_proto_init()
	file_diode_v1_interface_proto_init()
	file_diode_v1_manufacturer_proto_init()
	file_diode_v1_platform_proto_init()
	file_diode_v1_role_proto_init()
	file_diode_v1_site_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_diode_v1_ingester_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Entity); i {
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
		file_diode_v1_ingester_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IngestRequest); i {
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
		file_diode_v1_ingester_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IngestResponse); i {
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
	file_diode_v1_ingester_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Entity_Site)(nil),
		(*Entity_Platform)(nil),
		(*Entity_Manufacturer)(nil),
		(*Entity_Device)(nil),
		(*Entity_DeviceRole)(nil),
		(*Entity_DeviceType)(nil),
		(*Entity_Interface)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_diode_v1_ingester_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_diode_v1_ingester_proto_goTypes,
		DependencyIndexes: file_diode_v1_ingester_proto_depIdxs,
		MessageInfos:      file_diode_v1_ingester_proto_msgTypes,
	}.Build()
	File_diode_v1_ingester_proto = out.File
	file_diode_v1_ingester_proto_rawDesc = nil
	file_diode_v1_ingester_proto_goTypes = nil
	file_diode_v1_ingester_proto_depIdxs = nil
}
