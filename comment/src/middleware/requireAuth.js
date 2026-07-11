const jwt = require("jsonwebtoken");
const config = require("../config");

function requireAuth(req, res, next) {
  const bearer = req.header("Authorization") || "";
  const prefix = "Bearer ";
  if (!bearer.startsWith(prefix)) {
    return res.status(401).send("Invalid user id");
  }
  const token = bearer.slice(prefix.length);

  try {
    const payload = jwt.verify(token, config.jwtSecret);
    if (!payload.sub) {
      return res.status(401).send("Invalid user id");
    }
    req.userId = payload.sub;
    next();
  } catch (err) {
    return res.status(401).send("Invalid user id");
  }
}

module.exports = requireAuth;
