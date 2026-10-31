import { QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
import { ApiClientProvider } from "../api/ApiClientProvider.tsx";
import { createApiClient } from "../api/client.ts";
import { createQueryClient } from "../api/queryClient.ts";
import { createAppRouter } from "./router.tsx";

const queryClient = createQueryClient();
const apiClient = createApiClient();
const router = createAppRouter();

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ApiClientProvider client={apiClient}>
        <RouterProvider router={router} />
      </ApiClientProvider>
    </QueryClientProvider>
  );
}
