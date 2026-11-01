import { mergeConfig } from "vite";
import config from "./vite.config.ts";

// Isolated browser QA: the development database/API at 8080 is never mutated.
export default mergeConfig(config, {
  server: {
    port: 15173,
    proxy: {
      "/api": "http://127.0.0.1:18080",
      "/health": "http://127.0.0.1:18080",
    },
  },
});
