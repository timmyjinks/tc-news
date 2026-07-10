const axios = require("axios");
const config = require("./config");

async function postExists(postId) {
  try {
    await axios.get(
      `${config.postServiceUrl}/posts/${postId}`,
      { timeout: config.postServiceTimeout }
    );
    return true;
  } catch (err) {

    if (err.response?.status === 404)
      return false;

    throw err;
  }
}

module.exports = {
  postExists,
};
