/// <reference types="vitest/config" />
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// Where the dev and preview servers send API requests. Set in infra/compose.dev.yaml.
const backendUrl = process.env.BACKEND_URL ?? "http://localhost:8080";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    // Listen on all interfaces so the dev server is reachable from outside the Dev Container.
    host: true,
    port: 5173,
    strictPort: true,
    // The browser sees the frontend and the API on one origin, as in production
    // (ADR-0018), so the session cookie and the Origin check work the same way.
    proxy: {
      "/api": backendUrl,
      "/health": backendUrl,
    },
  },
  test: {
    // Components render into jsdom, so tests run without a browser.
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
  },
});
