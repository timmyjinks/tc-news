from datetime import datetime

from pydantic import BaseModel, ConfigDict, Field

DEFAULT_PARENT_ID = "00000000-0000-0000-0000-000000000000"


class CommentCreateRequest(BaseModel):
    body: str = ""
    parent_id: str = Field(default=DEFAULT_PARENT_ID)


class CommentUpdateRequest(BaseModel):
    body: str = ""


class CommentResponse(BaseModel):
    model_config = ConfigDict(populate_by_name=True)

    Id: str
    ParentId: str
    PostId: str
    UserId: str
    body: str
    created_at: datetime
