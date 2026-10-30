import { describe, expect, it, vi } from "vitest";
import { ApiError } from "./ApiError.ts";
import { problem } from "./fakeApiClient.ts";
import { createQueryClient } from "./queryClient.ts";

describe("createQueryClient", () => {
  describe("after a 401", () => {
    it("clears every query when a query fails", async () => {
      const queryClient = createQueryClient();
      queryClient.setQueryData(["previous-user"], "visible to the previous user");

      await queryClient
        .fetchQuery({
          queryKey: ["me"],
          queryFn: () => Promise.reject(problem(401, "unauthenticated")),
        })
        .catch(() => {});

      expect(queryClient.getQueryCache().getAll()).toHaveLength(0);
    });

    it("clears every query when a mutation fails", async () => {
      const queryClient = createQueryClient();
      queryClient.setQueryData(["previous-user"], "visible to the previous user");

      await queryClient
        .getMutationCache()
        .build(queryClient, { mutationFn: () => Promise.reject(problem(401, "unauthenticated")) })
        .execute(undefined)
        .catch(() => {});

      expect(queryClient.getQueryCache().getAll()).toHaveLength(0);
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
