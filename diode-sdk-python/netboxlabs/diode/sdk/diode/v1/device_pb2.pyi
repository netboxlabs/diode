from netboxlabs.diode.sdk.diode.v1 import device_type_pb2 as _device_type_pb2
from netboxlabs.diode.sdk.diode.v1 import platform_pb2 as _platform_pb2
from netboxlabs.diode.sdk.diode.v1 import role_pb2 as _role_pb2
from netboxlabs.diode.sdk.diode.v1 import site_pb2 as _site_pb2
from netboxlabs.diode.sdk.diode.v1 import tag_pb2 as _tag_pb2
from netboxlabs.diode.sdk.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Device(_message.Message):
    __slots__ = ("name", "device_fqdn", "device_type", "role", "platform", "serial", "site", "asset_tag", "status", "description", "comments", "tags")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DEVICE_FQDN_FIELD_NUMBER: _ClassVar[int]
    DEVICE_TYPE_FIELD_NUMBER: _ClassVar[int]
    ROLE_FIELD_NUMBER: _ClassVar[int]
    PLATFORM_FIELD_NUMBER: _ClassVar[int]
    SERIAL_FIELD_NUMBER: _ClassVar[int]
    SITE_FIELD_NUMBER: _ClassVar[int]
    ASSET_TAG_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    name: str
    device_fqdn: str
    device_type: _device_type_pb2.DeviceType
    role: _role_pb2.Role
    platform: _platform_pb2.Platform
    serial: str
    site: _site_pb2.Site
    asset_tag: str
    status: str
    description: str
    comments: str
    tags: _containers.RepeatedCompositeFieldContainer[_tag_pb2.Tag]
    def __init__(self, name: _Optional[str] = ..., device_fqdn: _Optional[str] = ..., device_type: _Optional[_Union[_device_type_pb2.DeviceType, _Mapping]] = ..., role: _Optional[_Union[_role_pb2.Role, _Mapping]] = ..., platform: _Optional[_Union[_platform_pb2.Platform, _Mapping]] = ..., serial: _Optional[str] = ..., site: _Optional[_Union[_site_pb2.Site, _Mapping]] = ..., asset_tag: _Optional[str] = ..., status: _Optional[str] = ..., description: _Optional[str] = ..., comments: _Optional[str] = ..., tags: _Optional[_Iterable[_Union[_tag_pb2.Tag, _Mapping]]] = ...) -> None: ...
