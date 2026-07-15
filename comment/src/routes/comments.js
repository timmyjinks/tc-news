const router = require("express").Router();

const { pool } = require("../db");
const kafka = require("../kafka");
const postClient = require("../postClient");
const requireAuth = require("../middleware/requireAuth");
const ZERO_UUID = "00000000-0000-0000-0000-000000000000";

router.get("/comments/:commentId", async (req, res, next) => {
  try {
    const { rows } = await pool.query(
      "SELECT * FROM comments WHERE id=$1",
      [req.params.commentId]
    );

    if (!rows.length)
      return res.sendStatus(404);

    res.json(rows[0]);

  } catch (err) {
    next(err);
  }
});

router.get("/posts/:postId/comments", async (req, res, next) => {
  try {
    const { rows } = await pool.query(
      "SELECT * FROM comments WHERE post_id=$1",
      [req.params.postId]
    );

    res.json(rows);

  } catch (err) {
    next(err);
  }
});

router.post("/posts/:postId/comments", requireAuth, async (req, res, next) => {

  try {

    const userId = req.userId;
    const parentId = req.body.parent_id ?? ZERO_UUID;

    const exists = await postClient.postExists(req.params.postId);
    if (!exists)
      return res.status(404).send("Post does not exist");

    const { rows } = await pool.query(`
            INSERT INTO comments(parent_id,post_id,user_id,body)
            VALUES($1,$2,$3,$4)
            RETURNING *
        `, [parentId, req.params.postId, userId, req.body.body]);

    const created = rows[0];

    if (parentId !== ZERO_UUID) {
      const parent = await pool.query(
        "SELECT user_id FROM comments WHERE id=$1",
        [parentId]
      );

      if (parent.rows.length) {
        await kafka.publishCommentReply({
          comment_id: created.id,
          post_id: created.post_id,
          user_id: created.user_id,
          parent_comment_id: parentId,
          parent_author_id: parent.rows[0].user_id,
          body: created.body,
        });
      }
    } else {
      await kafka.publishComment({
        comment_id: created.id,
        post_id: created.post_id,
        user_id: created.user_id,
        body: created.body,
      });
    }

    res.status(201).json(created);

  } catch (err) {
    next(err);
  }

});


router.put("/comments/:commentId", requireAuth, async (req, res, next) => {

  try {

    const userId = req.userId;

    const { rows } = await pool.query(`
            UPDATE comments
            SET body=$1
            WHERE id=$2
            AND user_id=$3
            RETURNING *
        `, [
      req.body.body,
      req.params.commentId,
      userId
    ]);

    if (!rows.length)
      return res.sendStatus(404);

    res.json(rows[0]);

  } catch (err) {
    next(err);
  }

});

router.delete("/comments/:commentId", requireAuth, async (req, res, next) => {

  try {

    const userId = req.userId;

    const result = await pool.query(
      "DELETE FROM comments WHERE id=$1 AND user_id=$2",
      [req.params.commentId, userId]
    );

    if (!result.rowCount)
      return res.sendStatus(404);

    res.sendStatus(204);

  } catch (err) {
    next(err);
  }

});

module.exports = router;
