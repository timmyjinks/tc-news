import logging

import httpx

from . import config

logger = logging.getLogger("comment.post_client")


class PostServiceUnavailable(Exception):
    """Raised when the Post Service can't be reached or returns something
    we don't know how to interpret."""


async def post_exists(post_id: str) -> bool:
    """Cross-service validation: ask the Post Service whether post_id is a
    real, existing post before we let a comment attach to it.

    Returns True if the Post Service confirms the post exists (200),
    False if the Post Service confirms it does not (404). Raises
    PostServiceUnavailable for anything else (network errors, timeouts,
    unexpected status codes), so the caller can fail safely instead of
    silently trusting the client-provided post_id.
    """
    url = f"{config.settings.post_service_url}/posts/{post_id}"

    try:
        async with httpx.AsyncClient(timeout=config.settings.post_service_timeout_seconds) as client:
            resp = await client.get(url)
    except httpx.RequestError as exc:
        logger.warning("[WARN] could not reach post service at %s: %s", url, exc)
        raise PostServiceUnavailable(str(exc)) from exc

    if resp.status_code == 200:
        return True
    if resp.status_code == 404:
        # post/cmd's GetPost handler returns a bare 404 for any lookup
        # error (row not found or otherwise), so this is the "does not
        # exist" signal to trust.
        return False

    logger.warning(
        "[WARN] unexpected status %s from post service for %s", resp.status_code, url
    )
    raise PostServiceUnavailable(f"unexpected status {resp.status_code} from post service")
