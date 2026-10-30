import { QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter } from "react-router";
import { ApiClientProvider } from "../api/ApiClientProvider.tsx";
import { createApiClient } from "../api/client.ts";
import { createQueryClient } from "../api/queryClient.ts";
import { AppRoutes } from "./routes.tsx";

const queryClient = createQueryClient();
const apiClient = createApiClient();

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ApiClientProvider client={apiClient}>
        <BrowserRouter>
          <AppRoutes />
        </BrowserRouter>
      </ApiClientProvider>
    </QueryClientProvider>
  );
}
