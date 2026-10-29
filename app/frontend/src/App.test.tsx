import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import App from "./App.tsx";

// A sample test for the test setup (#35). It goes away with the template App.
describe("App", () => {
  it("counts clicks", () => {
    render(<App />);

    const button = screen.getByRole("button", { name: "Count is 0" });
    fireEvent.click(button);

    expect(screen.getByRole("button", { name: "Count is 1" })).toBeInTheDocument();
  });
});
