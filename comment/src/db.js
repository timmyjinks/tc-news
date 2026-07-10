const { Pool } = require("pg");
const config = require("./config");

const pool = new Pool({
  connectionString: config.databaseUrl,
});

async function init() {
  await pool.query(`
        CREATE TABLE IF NOT EXISTS comments(
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            parent_id UUID DEFAULT '00000000-0000-0000-0000-000000000000',
            post_id UUID NOT NULL,
            user_id UUID NOT NULL,
            body TEXT DEFAULT '',
            created_at TIMESTAMP DEFAULT now()
        )
    `);
}

module.exports = {
  pool,
  init,
};
