// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

module.exports = function(api) {
  api.cache(true);
  return {
    presets: ["@babel/preset-env"],
    plugins: [["@babel/plugin-transform-runtime"]],
  };
};
