from netboxlabs.diode.sdk.diode.v1 import manufacturer_pb2 as _manufacturer_pb2
from netboxlabs.diode.sdk.validate import validate_pb2 as _validate_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DeviceType(_message.Message):
    __slots__ = ("manufacturer", "model", "slug")
    MANUFACTURER_FIELD_NUMBER: _ClassVar[int]
    MODEL_FIELD_NUMBER: _ClassVar[int]
    SLUG_FIELD_NUMBER: _ClassVar[int]
    manufacturer: _manufacturer_pb2.Manufacturer
    model: str
    slug: str
    def __init__(self, manufacturer: _Optional[_Union[_manufacturer_pb2.Manufacturer, _Mapping]] = ..., model: _Optional[str] = ..., slug: _Optional[str] = ...) -> None: ...
