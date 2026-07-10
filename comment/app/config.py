import os

from dotenv import load_dotenv

load_dotenv()


def _get_env(key: str, fallback: str) -> str:
    value = os.getenv(key)
    return value if value else fallback


def _build_db_uri() -> str:
    if uri := os.getenv("DATABASE_URL"):
        return uri

    host = _get_env("DB_HOST", "comment-db")
    port = _get_env("DB_PORT", "5432")
    user = _get_env("DB_USER", "postgres")
    password = _get_env("DB_PASSWORD", "password")
    name = _get_env("DB_NAME", "postgres")
    sslmode = _get_env("DB_SSLMODE", "disable")

    return f"postgres://{user}:{password}@{host}:{port}/{name}?sslmode={sslmode}"


class Settings:
    def __init__(self) -> None:
        self.database_url: str = _build_db_uri()
        self.db_pool_min_size: int = int(_get_env("DB_POOL_MIN_SIZE", "1"))
        self.db_pool_max_size: int = int(_get_env("DB_POOL_MAX_SIZE", "10"))

        self.kafka_broker: str = _get_env("KAFKA_BROKER", "kafka-service:9092")
        self.kafka_topic: str = _get_env("KAFKA_TOPIC", "notifications")

        self.host: str = _get_env("HOST", "0.0.0.0")
        self.port: int = int(_get_env("PORT", "8080"))

        self.log_level: str = _get_env("LOG_LEVEL", "INFO")

        self.post_service_url: str = _get_env(
            "POST_SERVICE_URL", "http://post:8080"
        ).rstrip("/")
        self.post_service_timeout_seconds: float = float(
            _get_env("POST_SERVICE_TIMEOUT_SECONDS", "5")
        )


settings = Settings()
