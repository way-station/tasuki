import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ApiError } from "../../../api/ApiError.ts";
import type { ApiClient } from "../../../api/client.ts";
import { createFakeApiClient } from "../../../api/fakeApiClient.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { HealthContainer } from "./HealthContainer.tsx";

function renderWith(client: ApiClient) {
  render(
    <TestProviders client={client}>
      <HealthContainer />
    </TestProviders>,
  );
}

describe("HealthContainer", () => {
  it("shows the health the backend returns", async () => {
    renderWith(
      createFakeApiClient({
        getHealth: () => Promise.resolve({ status: "error", mongo: "ok", postgres: "unreachable" }),
      }),
    );

    expect(await screen.findByText("Unhealthy")).toBeInTheDocument();
    expect(screen.getByText("PostgreSQL").nextSibling).toHaveTextContent("Unreachable");
  });

  it("shows that the backend cannot be reached when the request fails", async () => {
    renderWith(createFakeApiClient({ getHealth: () => Promise.reject(new ApiError(502)) }));

    expect(await screen.findByText("The backend cannot be reached.")).toBeInTheDocument();
  });
});
