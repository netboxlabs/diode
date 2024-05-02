from netboxlabs.diode.sdk.diode.v1 import site_pb2 as _site_pb2
from netboxlabs.diode.sdk.diode.v1 import tag_pb2 as _tag_pb2
from netboxlabs.diode.sdk.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Prefix(_message.Message):
    __slots__ = ("prefix", "site", "status", "is_pool", "mark_utilized", "description", "comments", "tags")
    PREFIX_FIELD_NUMBER: _ClassVar[int]
    SITE_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    IS_POOL_FIELD_NUMBER: _ClassVar[int]
    MARK_UTILIZED_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    prefix: str
    site: _site_pb2.Site
    status: str
    is_pool: bool
    mark_utilized: bool
    description: str
    comments: str
    tags: _containers.RepeatedCompositeFieldContainer[_tag_pb2.Tag]
    def __init__(self, prefix: _Optional[str] = ..., site: _Optional[_Union[_site_pb2.Site, _Mapping]] = ..., status: _Optional[str] = ..., is_pool: bool = ..., mark_utilized: bool = ..., description: _Optional[str] = ..., comments: _Optional[str] = ..., tags: _Optional[_Iterable[_Union[_tag_pb2.Tag, _Mapping]]] = ...) -> None: ...
