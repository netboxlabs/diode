from netboxlabs.diode.sdk.diode.v1 import device_type_pb2 as _device_type_pb2
from netboxlabs.diode.sdk.diode.v1 import platform_pb2 as _platform_pb2
from netboxlabs.diode.sdk.diode.v1 import role_pb2 as _role_pb2
from netboxlabs.diode.sdk.diode.v1 import site_pb2 as _site_pb2
from netboxlabs.diode.sdk.validate import validate_pb2 as _validate_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Device(_message.Message):
    __slots__ = ("name", "device_fqdn", "device_type", "role", "platform", "serial", "site", "vc_position")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DEVICE_FQDN_FIELD_NUMBER: _ClassVar[int]
    DEVICE_TYPE_FIELD_NUMBER: _ClassVar[int]
    ROLE_FIELD_NUMBER: _ClassVar[int]
    PLATFORM_FIELD_NUMBER: _ClassVar[int]
    SERIAL_FIELD_NUMBER: _ClassVar[int]
    SITE_FIELD_NUMBER: _ClassVar[int]
    VC_POSITION_FIELD_NUMBER: _ClassVar[int]
    name: str
    device_fqdn: str
    device_type: _device_type_pb2.DeviceType
    role: _role_pb2.Role
    platform: _platform_pb2.Platform
    serial: str
    site: _site_pb2.Site
    vc_position: int
    def __init__(self, name: _Optional[str] = ..., device_fqdn: _Optional[str] = ..., device_type: _Optional[_Union[_device_type_pb2.DeviceType, _Mapping]] = ..., role: _Optional[_Union[_role_pb2.Role, _Mapping]] = ..., platform: _Optional[_Union[_platform_pb2.Platform, _Mapping]] = ..., serial: _Optional[str] = ..., site: _Optional[_Union[_site_pb2.Site, _Mapping]] = ..., vc_position: _Optional[int] = ...) -> None: ...
