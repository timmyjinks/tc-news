import logging
from contextlib import asynccontextmanager

import asyncpg
from fastapi import FastAPI, Header, HTTPException, Path, Request, Response, status
from fastapi.responses import PlainTextResponse

from . import config, database, kafka_producer, post_client
from .models import CommentCreateRequest, CommentResponse, CommentUpdateRequest

logging.basicConfig(level=config.settings.log_level)
logger = logging.getLogger("comment")


@asynccontextmanager
async def lifespan(app: FastAPI):
    await database.init_db()
    await kafka_producer.start_producer()
    yield
    await kafka_producer.stop_producer()
    await database.close_db()


app = FastAPI(
    title="Comment Service API",
    description="API for creating, reading, updating, and deleting comments on posts.",
    version="3.0",
    lifespan=lifespan,
)


@app.exception_handler(HTTPException)
async def plain_text_http_exception_handler(request: Request, exc: HTTPException):
    return PlainTextResponse(str(exc.detail), status_code=exc.status_code)


def _row_to_comment(row: asyncpg.Record) -> CommentResponse:
    return CommentResponse(
        Id=str(row["id"]),
        ParentId=str(row["parent_id"]),
        PostId=str(row["post_id"]),
        UserId=str(row["user_id"]),
        body=row["body"],
        created_at=row["created_at"],
    )


def _affected_rows(command_tag: str) -> int:
    """asyncpg's Connection.execute returns a tag like 'UPDATE 1' or
    'DELETE 0'; pull the row count back out of it."""
    try:
        return int(command_tag.split()[-1])
    except (ValueError, IndexError):
        return 0


@app.get("/comments/{comment_id}", response_model=CommentResponse)
async def get_comment(comment_id: str = Path(...)):
    """Retrieves a single comment by its ID."""
    if not comment_id:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Comment does not exist")

    pool = database.get_pool()
    try:
        async with pool.acquire() as conn:
            row = await conn.fetchrow("SELECT * FROM comments WHERE id = $1", comment_id)
    except Exception as exc:
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=str(exc))

    if row is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Comment does not exist")

    return _row_to_comment(row)


@app.get("/posts/{post_id}/comments", response_model=list[CommentResponse])
async def list_comments(post_id: str = Path(...)):
    """Retrieves all comments belonging to a post."""
    if not post_id:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Post does not exist")

    pool = database.get_pool()
    try:
        async with pool.acquire() as conn:
            rows = await conn.fetch("SELECT * FROM comments WHERE post_id = $1", post_id)
    except Exception as exc:
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=str(exc))

    return [_row_to_comment(row) for row in rows]


@app.post(
    "/posts/{post_id}/comments",
    response_model=CommentResponse,
    status_code=status.HTTP_201_CREATED,
)
async def create_comment(
    post_id: str,
    comment: CommentCreateRequest,
    x_user_id: str | None = Header(default=None, alias="X-User-ID"),
):
    """Creates a new comment (optionally a reply, via parent_id) on a post.

    Stretch requirement: before adding the comment, we don't just trust the
    client's post_id -- we call the Post Service (POST_SERVICE_URL) to
    confirm the post actually exists. Missing post -> 404. Post Service
    unreachable/misbehaving -> 422, since we can't verify the client's
    input and won't silently accept it.
    """
    if not x_user_id:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid user id")
    if not post_id:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Post does not exist")

    try:
        exists = await post_client.post_exists(post_id)
    except post_client.PostServiceUnavailable as exc:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"could not verify post {post_id} with the post service: {exc}",
        )

    if not exists:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Post does not exist")

    pool = database.get_pool()
    try:
        async with pool.acquire() as conn:
            row = await conn.fetchrow(
                """
                INSERT INTO comments (parent_id, post_id, user_id, body)
                VALUES ($1, $2, $3, $4)
                RETURNING *
                """,
                comment.parent_id,
                post_id,
                x_user_id,
                comment.body,
            )
    except Exception as exc:
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=str(exc))

    await kafka_producer.send_comment_created(
        comment_id=str(row["id"]),
        post_id=post_id,
        user_id=x_user_id,
        body=comment.body,
    )

    return _row_to_comment(row)


@app.put("/comments/{comment_id}", response_model=CommentResponse)
async def update_comment(
    comment_id: str,
    comment: CommentUpdateRequest,
    x_user_id: str | None = Header(default=None, alias="X-User-ID"),
):
    """Updates the body of an existing comment."""
    if not comment_id:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Comment does not exist")
    if not x_user_id:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid user id")

    pool = database.get_pool()
    try:
        async with pool.acquire() as conn:
            tag = await conn.execute(
                "UPDATE comments SET body = $1 WHERE id = $2 AND user_id = $3",
                comment.body,
                comment_id,
                x_user_id,
            )
            if _affected_rows(tag) == 0:
                raise HTTPException(
                    status_code=status.HTTP_404_NOT_FOUND,
                    detail="Comment does not exist or invalid user id",
                )
            row = await conn.fetchrow("SELECT * FROM comments WHERE id = $1", comment_id)
    except HTTPException:
        raise
    except Exception as exc:
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=str(exc))

    return _row_to_comment(row)


@app.delete("/comments/{comment_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_comment(
    comment_id: str,
    x_user_id: str | None = Header(default=None, alias="X-User-ID"),
):
    """Deletes a comment owned by the authenticated user."""
    if not comment_id:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="Comment does not exist")
    if not x_user_id:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid user id")

    pool = database.get_pool()
    try:
        async with pool.acquire() as conn:
            tag = await conn.execute(
                "DELETE FROM comments WHERE id = $1 AND user_id = $2",
                comment_id,
                x_user_id,
            )
    except Exception as exc:
        raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=str(exc))

    if _affected_rows(tag) == 0:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Comment does not exist or invalid user id",
        )

    return Response(status_code=status.HTTP_204_NO_CONTENT)

if __name__ == "__main__":
    import uvicorn

    uvicorn.run("app.main:app", host=config.settings.host, port=config.settings.port)
