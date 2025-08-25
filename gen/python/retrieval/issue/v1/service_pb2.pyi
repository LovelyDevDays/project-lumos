from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RetrieveRequest(_message.Message):
    __slots__ = ("issue_keys",)
    ISSUE_KEYS_FIELD_NUMBER: _ClassVar[int]
    issue_keys: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, issue_keys: _Optional[_Iterable[str]] = ...) -> None: ...

class RetrieveResponse(_message.Message):
    __slots__ = ("issues",)
    ISSUES_FIELD_NUMBER: _ClassVar[int]
    issues: _containers.RepeatedCompositeFieldContainer[Issue]
    def __init__(self, issues: _Optional[_Iterable[_Union[Issue, _Mapping]]] = ...) -> None: ...

class Issue(_message.Message):
    __slots__ = ("key", "title", "content", "comments")
    KEY_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    key: str
    title: str
    content: str
    comments: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, key: _Optional[str] = ..., title: _Optional[str] = ..., content: _Optional[str] = ..., comments: _Optional[_Iterable[str]] = ...) -> None: ...
