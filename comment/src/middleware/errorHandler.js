const app = require("./app");
const db = require("./db");
const kafka = require("./kafka");
const config = require("./config");

async function main() {

  await db.init();
  await kafka.connect();

  app.listen(config.port, config.host, () => {
    console.log(`Listening on ${config.host}:${config.port}`);
  });

}

main();
