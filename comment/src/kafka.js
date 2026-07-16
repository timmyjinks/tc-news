const { Kafka } = require("kafkajs");
const config = require("./config");

const kafka = new Kafka({
  brokers: [config.kafkaBroker],
});

const producer = kafka.producer();

async function connect() {
  try {
    await producer.connect();
  } catch (err) {
    console.warn(err);
  }
}

async function publishComment(comment) {
  await producer.send({
    topic: config.kafkaTopic,
    messages: [{
      value: JSON.stringify({
        id: "",
        type: "comment_created",
        payload: comment,
      }),
    }],
  });
  await producer.send({
    topic: "events",
    messages: [{
      value: JSON.stringify({
        id: "",
        type: "comment_created",
        payload: comment,
      }),
    }],
  });
}

async function publishCommentReply(reply) {
  await producer.send({
    topic: config.kafkaTopic,
    messages: [{
      value: JSON.stringify({
        id: "",
        type: "comment_reply_created",
        payload: reply,
      }),
    }],
  });
  await producer.send({
    topic: "events",
    messages: [{
      value: JSON.stringify({
        id: "",
        type: "comment_reply_created",
        payload: reply,
      }),
    }],
  });
}

module.exports = {
  connect,
  publishComment,
  publishCommentReply,
};
