import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import DashboardPage from "../page";

vi.mock("@/store/auth-store", () => ({
  useAuthStore: () => ({
    user: {
      id: "user-123",
      email: "test@example.com",
    },
    isAuthenticated: true,
  }),
}));

describe("DashboardPage", () => {
  it("displays user email", () => {
    render(<DashboardPage />);
    expect(screen.getByText(/test@example\.com/)).toBeInTheDocument();
  });

  it("displays user ID", () => {
    render(<DashboardPage />);
    expect(screen.getByText(/user-123/)).toBeInTheDocument();
  });

  it("displays welcome heading", () => {
    render(<DashboardPage />);
    expect(
      screen.getByRole("heading", { name: /dashboard/i }),
    ).toBeInTheDocument();
  });
});
