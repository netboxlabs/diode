from netboxlabs.diode.sdk.validate import validate_pb2 as _validate_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Role(_message.Message):
    __slots__ = ("name", "slug", "vm_role")
    NAME_FIELD_NUMBER: _ClassVar[int]
    SLUG_FIELD_NUMBER: _ClassVar[int]
    VM_ROLE_FIELD_NUMBER: _ClassVar[int]
    name: str
    slug: str
    vm_role: bool
    def __init__(self, name: _Optional[str] = ..., slug: _Optional[str] = ..., vm_role: bool = ...) -> None: ...
