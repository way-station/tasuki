import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { HealthStatus } from "./HealthStatus.tsx";

describe("HealthStatus", () => {
  it("shows that the health is being checked", () => {
    render(<HealthStatus state="loading" />);

    expect(screen.getByText("Checking…")).toBeInTheDocument();
  });

  it("shows that the backend cannot be reached", () => {
    render(<HealthStatus state="unreachable" />);

    expect(screen.getByText("The backend cannot be reached.")).toBeInTheDocument();
  });

  it("shows each store", () => {
    render(
      <HealthStatus
        state="loaded"
        health={{
          healthy: false,
          stores: [
            { name: "MongoDB", reachable: true },
            { name: "PostgreSQL", reachable: false },
          ],
        }}
      />,
    );

    expect(screen.getByText("Unhealthy")).toBeInTheDocument();
    expect(screen.getByText("MongoDB").nextSibling).toHaveTextContent("Reachable");
    expect(screen.getByText("PostgreSQL").nextSibling).toHaveTextContent("Unreachable");
  });
});
