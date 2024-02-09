# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: diode/v1/distributor.proto
# Protobuf Python Version: 4.25.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder

# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()

DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(
    b'\n\x1a\x64iode/v1/distributor.proto\x12\x08\x64iode.v1\x1a\x15\x64iode/v1/device.proto\x1a\x1a\x64iode/v1/device_role.proto\x1a\x1a\x64iode/v1/device_type.proto\x1a\x18\x64iode/v1/interface.proto\x1a\x1b\x64iode/v1/manufacturer.proto\x1a\x17\x64iode/v1/platform.proto\x1a\x13\x64iode/v1/site.proto\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\x17validate/validate.proto\"\xc5\x03\n\x0cIngestEntity\x12$\n\x04site\x18\x01 \x01(\x0b\x32\x0e.diode.v1.SiteH\x00R\x04site\x12\x30\n\x08platform\x18\x02 \x01(\x0b\x32\x12.diode.v1.PlatformH\x00R\x08platform\x12<\n\x0cmanufacturer\x18\x03 \x01(\x0b\x32\x16.diode.v1.ManufacturerH\x00R\x0cmanufacturer\x12*\n\x06\x64\x65vice\x18\x04 \x01(\x0b\x32\x10.diode.v1.DeviceH\x00R\x06\x64\x65vice\x12\x37\n\x0b\x64\x65vice_role\x18\x05 \x01(\x0b\x32\x14.diode.v1.DeviceRoleH\x00R\ndeviceRole\x12\x37\n\x0b\x64\x65vice_type\x18\x06 \x01(\x0b\x32\x14.diode.v1.DeviceTypeH\x00R\ndeviceType\x12\x33\n\tinterface\x18\x07 \x01(\x0b\x32\x13.diode.v1.InterfaceH\x00R\tinterface\x12\x44\n\ttimestamp\x18\x08 \x01(\x0b\x32\x1a.google.protobuf.TimestampB\n\xfa\x42\x07\xb2\x01\x04\x08\x01\x38\x01R\ttimestampB\x06\n\x04\x64\x61ta\"\xe0\x02\n\x0bPushRequest\x12\"\n\x06stream\x18\x01 \x01(\tB\n\xfa\x42\x07r\x05\x10\x01\x18\xff\x01R\x06stream\x12\x37\n\x04\x64\x61ta\x18\x02 \x03(\x0b\x32\x16.diode.v1.IngestEntityB\x0b\xfa\x42\x08\x92\x01\x05\x08\x01\x10\xe8\x07R\x04\x64\x61ta\x12\x18\n\x02id\x18\x03 \x01(\tB\x08\xfa\x42\x05r\x03\xb0\x01\x01R\x02id\x12\x36\n\x11producer_app_name\x18\x04 \x01(\tB\n\xfa\x42\x07r\x05\x10\x01\x18\xff\x01R\x0fproducerAppName\x12<\n\x14producer_app_version\x18\x05 \x01(\tB\n\xfa\x42\x07r\x05\x10\x01\x18\xff\x01R\x12producerAppVersion\x12%\n\x08sdk_name\x18\x06 \x01(\tB\n\xfa\x42\x07r\x05\x10\x01\x18\xff\x01R\x07sdkName\x12=\n\x0bsdk_version\x18\x07 \x01(\tB\x1c\xfa\x42\x19r\x17\x32\x15^(\\d)+\\.(\\d)+\\.(\\d)+$R\nsdkVersion\"&\n\x0cPushResponse\x12\x16\n\x06\x65rrors\x18\x01 \x03(\tR\x06\x65rrors2M\n\x12\x44istributorService\x12\x37\n\x04Push\x12\x15.diode.v1.PushRequest\x1a\x16.diode.v1.PushResponse\"\x00\x42;Z9github.com/netboxlabs/diode/diode-sdk-go/diode/v1/diodepbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'diode.v1.distributor_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
    _globals['DESCRIPTOR']._options = None
    _globals['DESCRIPTOR']._serialized_options = b'Z9github.com/netboxlabs/diode/diode-sdk-go/diode/v1/diodepb'
    _globals['_INGESTENTITY'].fields_by_name['timestamp']._options = None
    _globals['_INGESTENTITY'].fields_by_name['timestamp']._serialized_options = b'\372B\007\262\001\004\010\0018\001'
    _globals['_PUSHREQUEST'].fields_by_name['stream']._options = None
    _globals['_PUSHREQUEST'].fields_by_name['stream']._serialized_options = b'\372B\007r\005\020\001\030\377\001'
    _globals['_PUSHREQUEST'].fields_by_name['data']._options = None
    _globals['_PUSHREQUEST'].fields_by_name['data']._serialized_options = b'\372B\010\222\001\005\010\001\020\350\007'
    _globals['_PUSHREQUEST'].fields_by_name['id']._options = None
    _globals['_PUSHREQUEST'].fields_by_name['id']._serialized_options = b'\372B\005r\003\260\001\001'
    _globals['_PUSHREQUEST'].fields_by_name['producer_app_name']._options = None
    _globals['_PUSHREQUEST'].fields_by_name[
        'producer_app_name']._serialized_options = b'\372B\007r\005\020\001\030\377\001'
    _globals['_PUSHREQUEST'].fields_by_name['producer_app_version']._options = None
    _globals['_PUSHREQUEST'].fields_by_name[
        'producer_app_version']._serialized_options = b'\372B\007r\005\020\001\030\377\001'
    _globals['_PUSHREQUEST'].fields_by_name['sdk_name']._options = None
    _globals['_PUSHREQUEST'].fields_by_name['sdk_name']._serialized_options = b'\372B\007r\005\020\001\030\377\001'
    _globals['_PUSHREQUEST'].fields_by_name['sdk_version']._options = None
    _globals['_PUSHREQUEST'].fields_by_name[
        'sdk_version']._serialized_options = b'\372B\031r\0272\025^(\\d)+\\.(\\d)+\\.(\\d)+$'
    _globals['_INGESTENTITY']._serialized_start = 279
    _globals['_INGESTENTITY']._serialized_end = 732
    _globals['_PUSHREQUEST']._serialized_start = 735
    _globals['_PUSHREQUEST']._serialized_end = 1087
    _globals['_PUSHRESPONSE']._serialized_start = 1089
    _globals['_PUSHRESPONSE']._serialized_end = 1127
    _globals['_DISTRIBUTORSERVICE']._serialized_start = 1129
    _globals['_DISTRIBUTORSERVICE']._serialized_end = 1206
# @@protoc_insertion_point(module_scope)
