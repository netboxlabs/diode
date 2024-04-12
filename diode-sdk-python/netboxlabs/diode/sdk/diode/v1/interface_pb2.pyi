from google.protobuf import any_pb2 as _any_pb2
from netboxlabs.diode.sdk.validate import validate_pb2 as _validate_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Interface(_message.Message):
    __slots__ = ("device", "name", "type", "enabled", "mtu", "mac_address", "mgmt_only")
    DEVICE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    MTU_FIELD_NUMBER: _ClassVar[int]
    MAC_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    MGMT_ONLY_FIELD_NUMBER: _ClassVar[int]
    device: _any_pb2.Any
    name: str
    type: str
    enabled: bool
    mtu: int
    mac_address: str
    mgmt_only: bool
    def __init__(self, device: _Optional[_Union[_any_pb2.Any, _Mapping]] = ..., name: _Optional[str] = ..., type: _Optional[str] = ..., enabled: bool = ..., mtu: _Optional[int] = ..., mac_address: _Optional[str] = ..., mgmt_only: bool = ...) -> None: ...
