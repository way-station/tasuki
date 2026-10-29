/// <reference types="vitest/config" />
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    // Listen on all interfaces so the dev server is reachable from outside the Dev Container.
    host: true,
    port: 5173,
    strictPort: true,
  },
  test: {
    // Components render into jsdom, so tests run without a browser.
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
  },
});
