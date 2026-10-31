import type { components } from "../../../api/schema.gen.ts";

// What the backend returns from GET /health.
export type HealthResponse = components["schemas"]["Health"];

// What the screen shows.
export type HealthView = {
  healthy: boolean;
  stores: StoreView[];
};

export type StoreView = {
  name: string;
  reachable: boolean;
};
