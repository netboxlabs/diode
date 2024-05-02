from netboxlabs.diode.sdk.diode.v1 import interface_pb2 as _interface_pb2
from netboxlabs.diode.sdk.diode.v1 import tag_pb2 as _tag_pb2
from netboxlabs.diode.sdk.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class IPAddress(_message.Message):
    __slots__ = ("address", "interface", "status", "role", "dns_name", "description", "comments", "tags")
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    INTERFACE_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ROLE_FIELD_NUMBER: _ClassVar[int]
    DNS_NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    address: str
    interface: _interface_pb2.Interface
    status: str
    role: str
    dns_name: str
    description: str
    comments: str
    tags: _containers.RepeatedCompositeFieldContainer[_tag_pb2.Tag]
    def __init__(self, address: _Optional[str] = ..., interface: _Optional[_Union[_interface_pb2.Interface, _Mapping]] = ..., status: _Optional[str] = ..., role: _Optional[str] = ..., dns_name: _Optional[str] = ..., description: _Optional[str] = ..., comments: _Optional[str] = ..., tags: _Optional[_Iterable[_Union[_tag_pb2.Tag, _Mapping]]] = ...) -> None: ...
