# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: diode/v1/ip_address.proto
# Protobuf Python Version: 5.26.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from netboxlabs.diode.sdk.diode.v1 import interface_pb2 as diode_dot_v1_dot_interface__pb2
from netboxlabs.diode.sdk.diode.v1 import tag_pb2 as diode_dot_v1_dot_tag__pb2
from netboxlabs.diode.sdk.validate import validate_pb2 as validate_dot_validate__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x19\x64iode/v1/ip_address.proto\x12\x08\x64iode.v1\x1a\x18\x64iode/v1/interface.proto\x1a\x12\x64iode/v1/tag.proto\x1a\x17validate/validate.proto\"X\n\x0e\x41ssignedObject\x12\x33\n\tinterface\x18\x01 \x01(\x0b\x32\x13.diode.v1.InterfaceH\x00R\tinterfaceB\x11\n\x0f\x61ssigned_object\"\xce\x03\n\tIPAddress\x12!\n\x07\x61\x64\x64ress\x18\x01 \x01(\tB\x07\xfa\x42\x04r\x02p\x01R\x07\x61\x64\x64ress\x12\x41\n\x0f\x61ssigned_object\x18\x02 \x01(\x0b\x32\x18.diode.v1.AssignedObjectR\x0e\x61ssignedObject\x12H\n\x06status\x18\x03 \x01(\tB0\xfa\x42-r+R\x06\x61\x63tiveR\x08reservedR\ndeprecatedR\x04\x64hcpR\x05slaacR\x06status\x12T\n\x04role\x18\x04 \x01(\tB@\xfa\x42=r;R\x08loopbackR\tsecondaryR\x07\x61nycastR\x03vipR\x04vrrpR\x04hsrpR\x04glbpR\x04\x63\x61rpR\x04role\x12P\n\x08\x64ns_name\x18\x05 \x01(\tB5\xfa\x42\x32r0\x18\xff\x01\x32+^([0-9A-Za-z_-]+|\\*)(\\.[0-9A-Za-z_-]+)*\\.?$R\x07\x64nsName\x12*\n\x0b\x64\x65scription\x18\x06 \x01(\tB\x08\xfa\x42\x05r\x03\x18\xc8\x01R\x0b\x64\x65scription\x12\x1a\n\x08\x63omments\x18\x07 \x01(\tR\x08\x63omments\x12!\n\x04tags\x18\x08 \x03(\x0b\x32\r.diode.v1.TagR\x04tagsB;Z9github.com/netboxlabs/diode/diode-sdk-go/diode/v1/diodepbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'diode.v1.ip_address_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z9github.com/netboxlabs/diode/diode-sdk-go/diode/v1/diodepb'
  _globals['_IPADDRESS'].fields_by_name['address']._loaded_options = None
  _globals['_IPADDRESS'].fields_by_name['address']._serialized_options = b'\372B\004r\002p\001'
  _globals['_IPADDRESS'].fields_by_name['status']._loaded_options = None
  _globals['_IPADDRESS'].fields_by_name['status']._serialized_options = b'\372B-r+R\006activeR\010reservedR\ndeprecatedR\004dhcpR\005slaac'
  _globals['_IPADDRESS'].fields_by_name['role']._loaded_options = None
  _globals['_IPADDRESS'].fields_by_name['role']._serialized_options = b'\372B=r;R\010loopbackR\tsecondaryR\007anycastR\003vipR\004vrrpR\004hsrpR\004glbpR\004carp'
  _globals['_IPADDRESS'].fields_by_name['dns_name']._loaded_options = None
  _globals['_IPADDRESS'].fields_by_name['dns_name']._serialized_options = b'\372B2r0\030\377\0012+^([0-9A-Za-z_-]+|\\*)(\\.[0-9A-Za-z_-]+)*\\.?$'
  _globals['_IPADDRESS'].fields_by_name['description']._loaded_options = None
  _globals['_IPADDRESS'].fields_by_name['description']._serialized_options = b'\372B\005r\003\030\310\001'
  _globals['_ASSIGNEDOBJECT']._serialized_start=110
  _globals['_ASSIGNEDOBJECT']._serialized_end=198
  _globals['_IPADDRESS']._serialized_start=201
  _globals['_IPADDRESS']._serialized_end=663
# @@protoc_insertion_point(module_scope)
