from netboxlabs.diode.sdk.diode.v1 import device_pb2 as _device_pb2
from netboxlabs.diode.sdk.diode.v1 import tag_pb2 as _tag_pb2
from netboxlabs.diode.sdk.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Interface(_message.Message):
    __slots__ = ("device", "name", "label", "type", "enabled", "mtu", "mac_address", "speed", "wwn", "mgmt_only", "description", "mark_connected", "mode", "tags")
    DEVICE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    LABEL_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    MTU_FIELD_NUMBER: _ClassVar[int]
    MAC_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    SPEED_FIELD_NUMBER: _ClassVar[int]
    WWN_FIELD_NUMBER: _ClassVar[int]
    MGMT_ONLY_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    MARK_CONNECTED_FIELD_NUMBER: _ClassVar[int]
    MODE_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    device: _device_pb2.Device
    name: str
    label: str
    type: str
    enabled: bool
    mtu: int
    mac_address: str
    speed: int
    wwn: str
    mgmt_only: bool
    description: str
    mark_connected: bool
    mode: str
    tags: _containers.RepeatedCompositeFieldContainer[_tag_pb2.Tag]
    def __init__(self, device: _Optional[_Union[_device_pb2.Device, _Mapping]] = ..., name: _Optional[str] = ..., label: _Optional[str] = ..., type: _Optional[str] = ..., enabled: bool = ..., mtu: _Optional[int] = ..., mac_address: _Optional[str] = ..., speed: _Optional[int] = ..., wwn: _Optional[str] = ..., mgmt_only: bool = ..., description: _Optional[str] = ..., mark_connected: bool = ..., mode: _Optional[str] = ..., tags: _Optional[_Iterable[_Union[_tag_pb2.Tag, _Mapping]]] = ...) -> None: ...
