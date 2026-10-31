import type { HealthResponse, HealthView } from "../types/health.ts";

export function toHealthView(health: HealthResponse): HealthView {
  return {
    healthy: health.status === "ok",
    stores: [
      { name: "MongoDB", reachable: health.mongo === "ok" },
      { name: "PostgreSQL", reachable: health.postgres === "ok" },
    ],
  };
}
