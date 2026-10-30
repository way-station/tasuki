import { createContext } from "react";
import type { ApiClient } from "./client.ts";

// Gives the ApiClient to the hooks. It passes a dependency, not state (ADR-0012).
export const ApiClientContext = createContext<ApiClient | null>(null);
