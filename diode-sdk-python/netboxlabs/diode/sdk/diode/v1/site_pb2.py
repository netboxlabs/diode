# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: diode/v1/site.proto
# Protobuf Python Version: 5.26.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from netboxlabs.diode.sdk.validate import validate_pb2 as validate_dot_validate__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x13\x64iode/v1/site.proto\x12\x08\x64iode.v1\x1a\x17validate/validate.proto\"V\n\x04Site\x12\x1d\n\x04name\x18\x01 \x01(\tB\t\xfa\x42\x06r\x04\x10\x01\x18\x64R\x04name\x12/\n\x04slug\x18\x02 \x01(\tB\x1b\xfa\x42\x18r\x16\x10\x01\x18\x64\x32\x10^[-a-zA-Z0-9_]+$R\x04slugB;Z9github.com/netboxlabs/diode/diode-sdk-go/diode/v1/diodepbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'diode.v1.site_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z9github.com/netboxlabs/diode/diode-sdk-go/diode/v1/diodepb'
  _globals['_SITE'].fields_by_name['name']._loaded_options = None
  _globals['_SITE'].fields_by_name['name']._serialized_options = b'\372B\006r\004\020\001\030d'
  _globals['_SITE'].fields_by_name['slug']._loaded_options = None
  _globals['_SITE'].fields_by_name['slug']._serialized_options = b'\372B\030r\026\020\001\030d2\020^[-a-zA-Z0-9_]+$'
  _globals['_SITE']._serialized_start=58
  _globals['_SITE']._serialized_end=144
# @@protoc_insertion_point(module_scope)
