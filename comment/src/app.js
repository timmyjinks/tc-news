const express = require("express");
const routes = require("./routes/comments");

const app = express();

app.use(express.json());

app.use(routes);

app.use((err, req, res, next) => {
  console.error(err);
  res.status(500).send(err.message);
});

module.exports = app;
