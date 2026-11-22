import { describe, it, expect, vi, beforeEach } from "vitest";
import { useAuthStore } from "@/store/auth-store";
import type { User } from "@/store/auth-store";
import * as api from "@/lib/api";

vi.mock("@/lib/api", () => ({
  apiClient: {
    post: vi.fn(),
    get: vi.fn(),
  },
}));

const mockUser: User = {
  id: "user-123",
  email: "test@example.com",
};

const mockAccessToken = "access-token-jwt";

describe("auth-store", () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: null,
      accessToken: null,
      isAuthenticated: false,
      isLoading: false,
    });
    vi.clearAllMocks();
  });

  describe("initial state", () => {
    it("has correct default values", () => {
      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.accessToken).toBeNull();
      expect(state.isAuthenticated).toBe(false);
      expect(state.isLoading).toBe(false);
    });
  });

  describe("login", () => {
    it("sets user, token, and isAuthenticated on success", async () => {
      vi.mocked(api.apiClient.post).mockResolvedValueOnce({
        data: {
          access_token: mockAccessToken,
          user: mockUser,
        },
      });

      await useAuthStore.getState().login("test@example.com", "password123");

      const state = useAuthStore.getState();
      expect(state.user).toEqual(mockUser);
      expect(state.accessToken).toBe(mockAccessToken);
      expect(state.isAuthenticated).toBe(true);
      expect(state.isLoading).toBe(false);
      expect(api.apiClient.post).toHaveBeenCalledWith("/api/auth/login", {
        email: "test@example.com",
        password: "password123",
      });
    });

    it("sets isLoading during login", async () => {
      let resolvePromise: (value: unknown) => void;
      const pendingPromise = new Promise((resolve) => {
        resolvePromise = resolve;
      });

      vi.mocked(api.apiClient.post).mockReturnValueOnce(
        pendingPromise as ReturnType<typeof api.apiClient.post>,
      );

      const loginPromise = useAuthStore
        .getState()
        .login("test@example.com", "password123");

      expect(useAuthStore.getState().isLoading).toBe(true);

      resolvePromise!({
        data: { access_token: mockAccessToken, user: mockUser },
      });

      await loginPromise;
      expect(useAuthStore.getState().isLoading).toBe(false);
    });

    it("resets state on login failure", async () => {
      vi.mocked(api.apiClient.post).mockRejectedValueOnce(
        new Error("Invalid credentials"),
      );

      await expect(
        useAuthStore.getState().login("test@example.com", "wrong"),
      ).rejects.toThrow("Invalid credentials");

      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.accessToken).toBeNull();
      expect(state.isAuthenticated).toBe(false);
      expect(state.isLoading).toBe(false);
    });
  });

  describe("register", () => {
    it("posts register with the real UserDTO shape, then auto-logs-in with real tokens", async () => {
      // Backend contract (spec user-auth.md): register returns 201 {id, email, created_at}
      vi.mocked(api.apiClient.post).mockResolvedValueOnce({
        data: mockUser, // UserDTO directly — NO access_token in this response
      });
      // auto-login response
      vi.mocked(api.apiClient.post).mockResolvedValueOnce({
        data: {
          access_token: mockAccessToken,
          user: mockUser,
        },
      });

      await useAuthStore
        .getState()
        .register("test@example.com", "password123");

      const state = useAuthStore.getState();
      expect(api.apiClient.post).toHaveBeenNthCalledWith(1, "/api/auth/register", {
        email: "test@example.com",
        password: "password123",
      });
      expect(api.apiClient.post).toHaveBeenNthCalledWith(2, "/api/auth/login", {
        email: "test@example.com",
        password: "password123",
      });
      expect(state.user).toEqual(mockUser);
      expect(state.accessToken).toBe(mockAccessToken);
      expect(state.isAuthenticated).toBe(true);
      expect(state.isLoading).toBe(false);
    });

    it("does not attempt login and clears state when register fails", async () => {
      vi.mocked(api.apiClient.post).mockRejectedValueOnce(
        new Error("Email already registered"),
      );

      await expect(
        useAuthStore.getState().register("test@example.com", "password123"),
      ).rejects.toThrow("Email already registered");

      expect(api.apiClient.post).toHaveBeenCalledTimes(1);
      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.isAuthenticated).toBe(false);
      expect(state.isLoading).toBe(false);
    });

    it("leaves user unauthenticated if auto-login fails after registration", async () => {
      vi.mocked(api.apiClient.post).mockResolvedValueOnce({ data: mockUser });
      vi.mocked(api.apiClient.post).mockRejectedValueOnce(new Error("login hiccup"));

      await expect(
        useAuthStore.getState().register("test@example.com", "password123"),
      ).rejects.toThrow("login hiccup");

      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(false);
      expect(state.accessToken).toBeNull();
    });
  });

  describe("logout", () => {
    it("clears state and calls logout endpoint", async () => {
      useAuthStore.setState({
        user: mockUser,
        accessToken: mockAccessToken,
        isAuthenticated: true,
      });

      vi.mocked(api.apiClient.post).mockResolvedValueOnce({ data: {} });

      await useAuthStore.getState().logout();

      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.accessToken).toBeNull();
      expect(state.isAuthenticated).toBe(false);
      expect(api.apiClient.post).toHaveBeenCalledWith("/api/auth/logout");
    });

    it("clears state even if logout API call fails", async () => {
      useAuthStore.setState({
        user: mockUser,
        accessToken: mockAccessToken,
        isAuthenticated: true,
      });

      vi.mocked(api.apiClient.post).mockRejectedValueOnce(
        new Error("Network error"),
      );

      await useAuthStore.getState().logout();

      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.accessToken).toBeNull();
      expect(state.isAuthenticated).toBe(false);
    });
  });

  describe("refreshToken", () => {
    it("updates accessToken on success", async () => {
      const newToken = "new-access-token";
      vi.mocked(api.apiClient.post).mockResolvedValueOnce({
        data: { access_token: newToken },
      });

      const result = await useAuthStore.getState().refreshToken();

      expect(result).toBe(newToken);
      expect(useAuthStore.getState().accessToken).toBe(newToken);
      expect(api.apiClient.post).toHaveBeenCalledWith("/api/auth/refresh");
    });

    it("returns null and clears state on failure", async () => {
      vi.mocked(api.apiClient.post).mockRejectedValueOnce(
        new Error("Refresh token expired"),
      );

      const result = await useAuthStore.getState().refreshToken();

      expect(result).toBeNull();
      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.accessToken).toBeNull();
      expect(state.isAuthenticated).toBe(false);
    });
  });
});
