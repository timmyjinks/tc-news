require("dotenv").config();

function env(key, fallback) {
  return process.env[key] || fallback;
}

function buildDbUrl() {
  if (process.env.DATABASE_URL)
    return process.env.DATABASE_URL;

  return `postgres://${env("DB_USER", "postgres")}:${env("DB_PASSWORD", "password")}@${env("DB_HOST", "comment-db")}:${env("DB_PORT", "5432")}/${env("DB_NAME", "postgres")}?sslmode=${env("DB_SSLMODE", "disable")}`;
}

module.exports = {
  databaseUrl: buildDbUrl(),
  kafkaBroker: env("KAFKA_BROKER", "kafka-service:9092"),
  kafkaTopic: env("KAFKA_TOPIC", "notifications"),

  host: env("HOST", "0.0.0.0"),
  port: Number(env("PORT", "8080")),

  postServiceUrl: env("POST_SERVICE_URL", "http://post:8080"),
  postServiceTimeout: Number(env("POST_SERVICE_TIMEOUT_SECONDS", "5")) * 1000,
};
