# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: diode/v1/device_role.proto
# Protobuf Python Version: 4.25.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder

# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()

DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(
    b'\n\x1a\x64iode/v1/device_role.proto\x12\x08\x64iode.v1\x1a\x17validate/validate.proto\"u\n\nDeviceRole\x12\x1d\n\x04name\x18\x01 \x01(\tB\t\xfa\x42\x06r\x04\x10\x01\x18\x64R\x04name\x12/\n\x04slug\x18\x02 \x01(\tB\x1b\xfa\x42\x18r\x16\x10\x01\x18\x64\x32\x10^[-a-zA-Z0-9_]+$R\x04slug\x12\x17\n\x07vm_role\x18\x03 \x01(\x08R\x06vmRoleBDZBgithub.com/netboxlabs/diode-internal/diode-sdk-go/diode/v1/diodepbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'diode.v1.device_role_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
    _globals['DESCRIPTOR']._options = None
    _globals['DESCRIPTOR']._serialized_options = b'ZBgithub.com/netboxlabs/diode-internal/diode-sdk-go/diode/v1/diodepb'
    _globals['_DEVICEROLE'].fields_by_name['name']._options = None
    _globals['_DEVICEROLE'].fields_by_name['name']._serialized_options = b'\372B\006r\004\020\001\030d'
    _globals['_DEVICEROLE'].fields_by_name['slug']._options = None
    _globals['_DEVICEROLE'].fields_by_name[
        'slug']._serialized_options = b'\372B\030r\026\020\001\030d2\020^[-a-zA-Z0-9_]+$'
    _globals['_DEVICEROLE']._serialized_start = 65
    _globals['_DEVICEROLE']._serialized_end = 182
# @@protoc_insertion_point(module_scope)
