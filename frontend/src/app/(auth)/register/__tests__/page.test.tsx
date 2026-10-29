import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import RegisterPage from "../page";

const mockPush = vi.fn();
const mockRegister = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}));

vi.mock("@/store/auth-store", () => ({
  useAuthStore: () => ({
    register: mockRegister,
    isLoading: false,
  }),
}));

describe("RegisterPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders register form with email, password, and confirm fields", () => {
    render(<RegisterPage />);
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getAllByLabelText(/password/i)).toHaveLength(2);
    expect(
      screen.getByRole("button", { name: /sign up/i }),
    ).toBeInTheDocument();
  });

  it("shows error when passwords do not match", async () => {
    const user = userEvent.setup();
    render(<RegisterPage />);

    await user.type(screen.getByLabelText(/email/i), "test@example.com");
    await user.type(screen.getAllByLabelText(/password/i)[0], "password123");
    await user.type(screen.getAllByLabelText(/password/i)[1], "different");
    await user.click(screen.getByRole("button", { name: /sign up/i }));

    expect(mockRegister).not.toHaveBeenCalled();
    expect(screen.getByText(/passwords do not match/i)).toBeInTheDocument();
  });

  it("calls store.register with email and password on submit", async () => {
    const user = userEvent.setup();
    mockRegister.mockResolvedValueOnce(undefined);
    render(<RegisterPage />);

    await user.type(screen.getByLabelText(/email/i), "test@example.com");
    await user.type(screen.getAllByLabelText(/password/i)[0], "password123");
    await user.type(screen.getAllByLabelText(/password/i)[1], "password123");
    await user.click(screen.getByRole("button", { name: /sign up/i }));

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith(
        "test@example.com",
        "password123",
      );
    });
  });

  it("redirects to /dashboard on successful registration (auto-login)", async () => {
    const user = userEvent.setup();
    mockRegister.mockResolvedValueOnce(undefined);
    render(<RegisterPage />);

    await user.type(screen.getByLabelText(/email/i), "test@example.com");
    await user.type(screen.getAllByLabelText(/password/i)[0], "password123");
    await user.type(screen.getAllByLabelText(/password/i)[1], "password123");
    await user.click(screen.getByRole("button", { name: /sign up/i }));

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith("/dashboard");
    });
  });

  it("displays error message on register failure", async () => {
    const user = userEvent.setup();
    mockRegister.mockRejectedValueOnce(
      new Error("Email already registered"),
    );
    render(<RegisterPage />);

    await user.type(screen.getByLabelText(/email/i), "test@example.com");
    await user.type(screen.getAllByLabelText(/password/i)[0], "password123");
    await user.type(screen.getAllByLabelText(/password/i)[1], "password123");
    await user.click(screen.getByRole("button", { name: /sign up/i }));

    await waitFor(() => {
      expect(
        screen.getByText(/email already registered/i),
      ).toBeInTheDocument();
    });
  });

  it("has link to login page", () => {
    render(<RegisterPage />);
    const loginLink = screen.getByRole("link", { name: /sign in/i });
    expect(loginLink).toHaveAttribute("href", "/login");
  });
});
