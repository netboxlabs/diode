from netboxlabs.diode.sdk.diode.v1 import device_pb2 as _device_pb2
from netboxlabs.diode.sdk.diode.v1 import device_type_pb2 as _device_type_pb2
from netboxlabs.diode.sdk.diode.v1 import interface_pb2 as _interface_pb2
from netboxlabs.diode.sdk.diode.v1 import ip_address_pb2 as _ip_address_pb2
from netboxlabs.diode.sdk.diode.v1 import manufacturer_pb2 as _manufacturer_pb2
from netboxlabs.diode.sdk.diode.v1 import platform_pb2 as _platform_pb2
from netboxlabs.diode.sdk.diode.v1 import role_pb2 as _role_pb2
from netboxlabs.diode.sdk.diode.v1 import site_pb2 as _site_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from netboxlabs.diode.sdk.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Entity(_message.Message):
    __slots__ = ("site", "platform", "manufacturer", "device", "device_role", "device_type", "interface", "ip_address", "timestamp")
    SITE_FIELD_NUMBER: _ClassVar[int]
    PLATFORM_FIELD_NUMBER: _ClassVar[int]
    MANUFACTURER_FIELD_NUMBER: _ClassVar[int]
    DEVICE_FIELD_NUMBER: _ClassVar[int]
    DEVICE_ROLE_FIELD_NUMBER: _ClassVar[int]
    DEVICE_TYPE_FIELD_NUMBER: _ClassVar[int]
    INTERFACE_FIELD_NUMBER: _ClassVar[int]
    IP_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    site: _site_pb2.Site
    platform: _platform_pb2.Platform
    manufacturer: _manufacturer_pb2.Manufacturer
    device: _device_pb2.Device
    device_role: _role_pb2.Role
    device_type: _device_type_pb2.DeviceType
    interface: _interface_pb2.Interface
    ip_address: _ip_address_pb2.IPAddress
    timestamp: _timestamp_pb2.Timestamp
    def __init__(self, site: _Optional[_Union[_site_pb2.Site, _Mapping]] = ..., platform: _Optional[_Union[_platform_pb2.Platform, _Mapping]] = ..., manufacturer: _Optional[_Union[_manufacturer_pb2.Manufacturer, _Mapping]] = ..., device: _Optional[_Union[_device_pb2.Device, _Mapping]] = ..., device_role: _Optional[_Union[_role_pb2.Role, _Mapping]] = ..., device_type: _Optional[_Union[_device_type_pb2.DeviceType, _Mapping]] = ..., interface: _Optional[_Union[_interface_pb2.Interface, _Mapping]] = ..., ip_address: _Optional[_Union[_ip_address_pb2.IPAddress, _Mapping]] = ..., timestamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class IngestRequest(_message.Message):
    __slots__ = ("stream", "entities", "id", "producer_app_name", "producer_app_version", "sdk_name", "sdk_version")
    STREAM_FIELD_NUMBER: _ClassVar[int]
    ENTITIES_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    PRODUCER_APP_NAME_FIELD_NUMBER: _ClassVar[int]
    PRODUCER_APP_VERSION_FIELD_NUMBER: _ClassVar[int]
    SDK_NAME_FIELD_NUMBER: _ClassVar[int]
    SDK_VERSION_FIELD_NUMBER: _ClassVar[int]
    stream: str
    entities: _containers.RepeatedCompositeFieldContainer[Entity]
    id: str
    producer_app_name: str
    producer_app_version: str
    sdk_name: str
    sdk_version: str
    def __init__(self, stream: _Optional[str] = ..., entities: _Optional[_Iterable[_Union[Entity, _Mapping]]] = ..., id: _Optional[str] = ..., producer_app_name: _Optional[str] = ..., producer_app_version: _Optional[str] = ..., sdk_name: _Optional[str] = ..., sdk_version: _Optional[str] = ...) -> None: ...

class IngestResponse(_message.Message):
    __slots__ = ("errors",)
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    errors: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, errors: _Optional[_Iterable[str]] = ...) -> None: ...
