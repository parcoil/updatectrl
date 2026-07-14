"use strict";

const { cleanup } = require("./install");

module.exports = { cleanup };

if (require.main === module) {
  cleanup()
    .then(() => process.exit(0))
    .catch(() => process.exit(0));
}
