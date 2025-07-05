from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class SendReplyCommand(_message.Message):
    __slots__ = ("conversation_id", "message")
    CONVERSATION_ID_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    conversation_id: str
    message: str
    def __init__(self, conversation_id: _Optional[str] = ..., message: _Optional[str] = ...) -> None: ...

class Status(_message.Message):
    __slots__ = ("success", "error")
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    success: bool
    error: str
    def __init__(self, success: bool = ..., error: _Optional[str] = ...) -> None: ...
