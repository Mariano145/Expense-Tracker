import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import Hello from "./Hello";

describe("Hello", () => {
  it("renders default greeting", () => {
    render(<Hello />);
    expect(screen.getByText("Hello, World!")).toBeInTheDocument();
  });

  it("renders custom name", () => {
    render(<Hello name="Spendly" />);
    expect(screen.getByText("Hello, Spendly!")).toBeInTheDocument();
  });
});
