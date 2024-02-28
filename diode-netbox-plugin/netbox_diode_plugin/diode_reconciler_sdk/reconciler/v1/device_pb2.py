# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: reconciler/v1/device.proto
# Protobuf Python Version: 4.25.1
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from netbox_diode_plugin.diode_reconciler_sdk.reconciler.v1 import device_role_pb2 as reconciler_dot_v1_dot_device__role__pb2
from netbox_diode_plugin.diode_reconciler_sdk.reconciler.v1 import device_type_pb2 as reconciler_dot_v1_dot_device__type__pb2
from netbox_diode_plugin.diode_reconciler_sdk.reconciler.v1 import platform_pb2 as reconciler_dot_v1_dot_platform__pb2
from netbox_diode_plugin.diode_reconciler_sdk.reconciler.v1 import site_pb2 as reconciler_dot_v1_dot_site__pb2
from netbox_diode_plugin.diode_reconciler_sdk.validate import validate_pb2 as validate_dot_validate__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1areconciler/v1/device.proto\x12\rreconciler.v1\x1a\x1freconciler/v1/device_role.proto\x1a\x1freconciler/v1/device_type.proto\x1a\x1creconciler/v1/platform.proto\x1a\x18reconciler/v1/site.proto\x1a\x17validate/validate.proto\"\xd5\x03\n\x06\x44\x65vice\x12\x17\n\x02id\x18\x01 \x01(\x04\x42\x07\xfa\x42\x04\x32\x02(\x01R\x02id\x12\x1b\n\x04name\x18\x02 \x01(\tB\x07\xfa\x42\x04r\x02\x18@R\x04name\x12\x44\n\x0b\x64\x65vice_type\x18\x03 \x01(\x0b\x32\x19.reconciler.v1.DeviceTypeB\x08\xfa\x42\x05\xa2\x01\x02\x08\x01R\ndeviceType\x12\x37\n\x04role\x18\x04 \x01(\x0b\x32\x19.reconciler.v1.DeviceRoleB\x08\xfa\x42\x05\xa2\x01\x02\x08\x01R\x04role\x12\x44\n\x0b\x64\x65vice_role\x18\x05 \x01(\x0b\x32\x19.reconciler.v1.DeviceRoleB\x08\xfa\x42\x05\xa2\x01\x02\x08\x01R\ndeviceRole\x12\x33\n\x08platform\x18\x06 \x01(\x0b\x32\x17.reconciler.v1.PlatformR\x08platform\x12\x1f\n\x06serial\x18\x07 \x01(\tB\x07\xfa\x42\x04r\x02\x18\x32R\x06serial\x12\x31\n\x04site\x18\x08 \x01(\x0b\x32\x13.reconciler.v1.SiteB\x08\xfa\x42\x05\xa2\x01\x02\x08\x01R\x04site\x12+\n\x0bvc_position\x18\t \x01(\x05\x42\n\xfa\x42\x07\x1a\x05\x18\xff\x01(\x00R\nvcPosition\x12\x1a\n\x03url\x18\n \x01(\tB\x08\xfa\x42\x05r\x03\x88\x01\x01R\x03urlBEZCgithub.com/netboxlabs/diode/diode-server/reconciler/v1/reconcilerpbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'reconciler.v1.device_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  _globals['DESCRIPTOR']._options = None
  _globals['DESCRIPTOR']._serialized_options = b'ZCgithub.com/netboxlabs/diode/diode-server/reconciler/v1/reconcilerpb'
  _globals['_DEVICE'].fields_by_name['id']._options = None
  _globals['_DEVICE'].fields_by_name['id']._serialized_options = b'\372B\0042\002(\001'
  _globals['_DEVICE'].fields_by_name['name']._options = None
  _globals['_DEVICE'].fields_by_name['name']._serialized_options = b'\372B\004r\002\030@'
  _globals['_DEVICE'].fields_by_name['device_type']._options = None
  _globals['_DEVICE'].fields_by_name['device_type']._serialized_options = b'\372B\005\242\001\002\010\001'
  _globals['_DEVICE'].fields_by_name['role']._options = None
  _globals['_DEVICE'].fields_by_name['role']._serialized_options = b'\372B\005\242\001\002\010\001'
  _globals['_DEVICE'].fields_by_name['device_role']._options = None
  _globals['_DEVICE'].fields_by_name['device_role']._serialized_options = b'\372B\005\242\001\002\010\001'
  _globals['_DEVICE'].fields_by_name['serial']._options = None
  _globals['_DEVICE'].fields_by_name['serial']._serialized_options = b'\372B\004r\002\0302'
  _globals['_DEVICE'].fields_by_name['site']._options = None
  _globals['_DEVICE'].fields_by_name['site']._serialized_options = b'\372B\005\242\001\002\010\001'
  _globals['_DEVICE'].fields_by_name['vc_position']._options = None
  _globals['_DEVICE'].fields_by_name['vc_position']._serialized_options = b'\372B\007\032\005\030\377\001(\000'
  _globals['_DEVICE'].fields_by_name['url']._options = None
  _globals['_DEVICE'].fields_by_name['url']._serialized_options = b'\372B\005r\003\210\001\001'
  _globals['_DEVICE']._serialized_start=193
  _globals['_DEVICE']._serialized_end=662
# @@protoc_insertion_point(module_scope)