import { QueryClientProvider, useQuery } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import { ApiError } from "./ApiError.ts";
import { problem } from "./fakeApiClient.ts";
import { createQueryClient } from "./queryClient.ts";

describe("createQueryClient", () => {
  describe("after a 401", () => {
    it("removes every query with data when a query fails", async () => {
      const queryClient = createQueryClient();
      queryClient.setQueryData(["previous-user"], "visible to the previous user");

      await queryClient
        .fetchQuery({
          queryKey: ["me"],
          queryFn: () => Promise.reject(problem(401, "unauthenticated")),
        })
        .catch(() => {});

      expect(queryClient.getQueryData(["previous-user"])).toBeUndefined();
    });

    it("removes every query with data when a mutation fails", async () => {
      const queryClient = createQueryClient();
      queryClient.setQueryData(["previous-user"], "visible to the previous user");

      await queryClient
        .getMutationCache()
        .build(queryClient, { mutationFn: () => Promise.reject(problem(401, "unauthenticated")) })
        .execute(undefined)
        .catch(() => {});

      expect(queryClient.getQueryData(["previous-user"])).toBeUndefined();
    });

    it("does not fetch a query in use again and again", async () => {
      const queryClient = createQueryClient();
      const queryFn = vi.fn(() => Promise.reject(problem(401, "unauthenticated")));
      const wrapper = ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
      );

      // A screen using the query.
      const { result } = renderHook(() => useQuery({ queryKey: ["me"], queryFn }), { wrapper });

      await waitFor(() => expect(result.current.isError).toBe(true));
      await new Promise((resolve) => setTimeout(resolve, 100));
      expect(queryFn).toHaveBeenCalledTimes(1);
    });

    it("removes the failing query's own data, fetching it only once more", async () => {
      const queryClient = createQueryClient();
      queryClient.setQueryData(["me"], "the previous user");
      const queryFn = vi.fn(() => Promise.reject(problem(401, "unauthenticated")));
      const wrapper = ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
      );

      const { result } = renderHook(() => useQuery({ queryKey: ["me"], queryFn }), { wrapper });

      await waitFor(() => expect(result.current.isError).toBe(true));
      await new Promise((resolve) => setTimeout(resolve, 100));
      expect(result.current.data).toBeUndefined();
      expect(queryFn).toHaveBeenCalledTimes(2);
    });

    it("keeps the queries after other errors", async () => {
      const queryClient = createQueryClient();
      queryClient.setQueryData(["race"], "a race");

      await queryClient
        .fetchQuery({
          queryKey: ["missing"],
          queryFn: () => Promise.reject(problem(404, "not_found")),
        })
        .catch(() => {});

      expect(queryClient.getQueryData(["race"])).toBe("a race");
    });
  });

  describe("retries", () => {
    it("does not retry a 4xx", async () => {
      const queryClient = createQueryClient();
      const queryFn = vi.fn(() => Promise.reject(problem(404, "not_found")));

      await queryClient
        .fetchQuery({ queryKey: ["missing"], queryFn, retryDelay: 0 })
        .catch(() => {});

      expect(queryFn).toHaveBeenCalledTimes(1);
    });

    it("retries a 5xx three times", async () => {
      const queryClient = createQueryClient();
      const queryFn = vi.fn(() => Promise.reject(new ApiError(502)));

      await queryClient.fetchQuery({ queryKey: ["flaky"], queryFn, retryDelay: 0 }).catch(() => {});

      expect(queryFn).toHaveBeenCalledTimes(4);
    });
  });
});
