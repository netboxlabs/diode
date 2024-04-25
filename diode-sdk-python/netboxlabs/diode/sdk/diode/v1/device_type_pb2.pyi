from netboxlabs.diode.sdk.diode.v1 import manufacturer_pb2 as _manufacturer_pb2
from netboxlabs.diode.sdk.diode.v1 import tag_pb2 as _tag_pb2
from netboxlabs.diode.sdk.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DeviceType(_message.Message):
    __slots__ = ("model", "slug", "manufacturer", "description", "comments", "part_number", "tags")
    MODEL_FIELD_NUMBER: _ClassVar[int]
    SLUG_FIELD_NUMBER: _ClassVar[int]
    MANUFACTURER_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    PART_NUMBER_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    model: str
    slug: str
    manufacturer: _manufacturer_pb2.Manufacturer
    description: str
    comments: str
    part_number: str
    tags: _containers.RepeatedCompositeFieldContainer[_tag_pb2.Tag]
    def __init__(self, model: _Optional[str] = ..., slug: _Optional[str] = ..., manufacturer: _Optional[_Union[_manufacturer_pb2.Manufacturer, _Mapping]] = ..., description: _Optional[str] = ..., comments: _Optional[str] = ..., part_number: _Optional[str] = ..., tags: _Optional[_Iterable[_Union[_tag_pb2.Tag, _Mapping]]] = ...) -> None: ...
