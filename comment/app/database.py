import os

import asyncpg


DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "postgres://postgres:password@comment-db:5432/postgres",
)

CREATE_TABLE_SQL = """
CREATE TABLE IF NOT EXISTS comments (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id  uuid DEFAULT '00000000-0000-0000-0000-000000000000'::uuid,
    post_id    uuid NOT NULL,
    user_id    uuid NOT NULL,
    body       TEXT DEFAULT '{}',
    created_at TIMESTAMP DEFAULT now()
)
"""

_pool: asyncpg.Pool | None = None


async def init_db() -> asyncpg.Pool:
    global _pool
    _pool = await asyncpg.create_pool(dsn=DATABASE_URL)
    async with _pool.acquire() as conn:
        await conn.execute(CREATE_TABLE_SQL)
    return _pool


async def close_db() -> None:
    global _pool
    if _pool is not None:
        await _pool.close()
        _pool = None


def get_pool() -> asyncpg.Pool:
    if _pool is None:
        raise RuntimeError("database pool has not been initialized")
    return _pool
