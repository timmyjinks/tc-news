import json
import logging
import os

from aiokafka import AIOKafkaProducer

logger = logging.getLogger("comment.kafka")

KAFKA_BROKER = os.getenv("KAFKA_BROKER", "kafka-service:9092")
TOPIC = "notifications"

_producer: AIOKafkaProducer | None = None


async def start_producer() -> None:
    global _producer
    _producer = AIOKafkaProducer(bootstrap_servers=KAFKA_BROKER)
    try:
        await _producer.start()
    except Exception as exc:  # pragma: no cover - matches Go's log.Println("[WARN]"...)
        logger.warning("[WARN] failed to connect to kafka: %s", exc)


async def stop_producer() -> None:
    global _producer
    if _producer is not None:
        await _producer.stop()
        _producer = None


async def send_comment_created(comment_id: str, post_id: str, user_id: str, body: str) -> None:
    if _producer is None:
        logger.warning("[WARN] kafka producer not initialized; skipping publish")
        return

    envelope = {
        "id": "",
        "type": "comment_created",
        "payload": {
            "comment_id": comment_id,
            "post_id": post_id,
            "user_id": user_id,
            "body": body,
        },
    }

    try:
        await _producer.send_and_wait(TOPIC, json.dumps(envelope).encode("utf-8"))
    except Exception as exc:
        # Matches: log.Println("[WARN] failed to publish comment_created event:", err)
        # The original handler never surfaces this to the HTTP caller, so
        # a publish failure does not fail the request.
        logger.warning("[WARN] failed to publish comment_created event: %s", exc)
