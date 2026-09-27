import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { InternalAxiosRequestConfig, AxiosResponse, AxiosError } from "axios";

let requestInterceptor: (config: InternalAxiosRequestConfig) => InternalAxiosRequestConfig;
let responseErrorInterceptor: (error: AxiosError) => Promise<unknown>;

function createHeadersMock() {
  const store = new Map<string, string>();
  return {
    set: vi.fn((key: string, value: string) => store.set(key, value)),
    get: vi.fn((key: string) => store.get(key)),
    has: vi.fn((key: string) => store.has(key)),
    delete: vi.fn((key: string) => store.delete(key)),
    common: {},
  };
}

const mockInstanceFn = vi.fn().mockResolvedValue({ data: {} });

vi.mock("axios", () => {
  const interceptors = {
    request: {
      use: vi.fn((fn: typeof requestInterceptor) => {
        requestInterceptor = fn;
        return 0;
      }),
    },
    response: {
      use: vi.fn(
        (_success: unknown, error: typeof responseErrorInterceptor) => {
          responseErrorInterceptor = error;
          return 0;
        },
      ),
    },
  };

  const instance = Object.assign(mockInstanceFn, {
    interceptors,
    defaults: { headers: { common: {} } },
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
    patch: vi.fn(),
    request: vi.fn(),
  });

  return {
    default: {
      create: vi.fn(() => instance),
    },
  };
});

vi.mock("@/store/auth-store", () => ({
  useAuthStore: {
    getState: vi.fn(() => ({
      accessToken: null,
      refreshToken: vi.fn().mockResolvedValue(null),
      logout: vi.fn().mockResolvedValue(undefined),
    })),
  },
}));

import { useAuthStore } from "@/store/auth-store";

describe("api client", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.unstubAllEnvs();
  });

  afterEach(() => {
    vi.resetModules();
  });

  describe("axios instance creation", () => {
    it("creates instance with baseURL from env and withCredentials", async () => {
      vi.stubEnv("NEXT_PUBLIC_API_URL", "http://test-api:9090");
      const axios = await import("axios");
      const { apiClient } = await import("@/lib/api");

      expect(axios.default.create).toHaveBeenCalledWith(
        expect.objectContaining({
          baseURL: "http://test-api:9090",
          withCredentials: true,
        }),
      );
      expect(apiClient).toBeDefined();
    });

    it("defaults baseURL to http://localhost:8080 when env not set", async () => {
      const axios = await import("axios");
      await import("@/lib/api");

      expect(axios.default.create).toHaveBeenCalledWith(
        expect.objectContaining({
          baseURL: "http://localhost:8080",
        }),
      );
    });
  });

  describe("request interceptor", () => {
    it("adds Bearer token when accessToken exists in store", async () => {
      vi.mocked(useAuthStore.getState).mockReturnValue({
        accessToken: "my-jwt-token",
        user: null,
        isAuthenticated: true,
        isLoading: false,
        login: vi.fn(),
        register: vi.fn(),
        logout: vi.fn(),
        refreshToken: vi.fn(),
      });

      await import("@/lib/api");

      const headers = createHeadersMock();
      const config = {
        headers,
      } as unknown as InternalAxiosRequestConfig;

      const result = requestInterceptor(config);
      expect(result.headers.set).toHaveBeenCalledWith(
        "Authorization",
        "Bearer my-jwt-token",
      );
    });

    it("does not add Authorization header when no token", async () => {
      vi.mocked(useAuthStore.getState).mockReturnValue({
        accessToken: null,
        user: null,
        isAuthenticated: false,
        isLoading: false,
        login: vi.fn(),
        register: vi.fn(),
        logout: vi.fn(),
        refreshToken: vi.fn(),
      });

      await import("@/lib/api");

      const headers = createHeadersMock();
      const config = {
        headers,
      } as unknown as InternalAxiosRequestConfig;

      requestInterceptor(config);
      expect(headers.set).not.toHaveBeenCalled();
    });
  });

  describe("response interceptor — 401 handling", () => {
    it("refreshes token and retries original request on 401", async () => {
      const refreshMock = vi.fn().mockResolvedValue("new-token");
      vi.mocked(useAuthStore.getState).mockReturnValue({
        accessToken: "expired-token",
        user: null,
        isAuthenticated: true,
        isLoading: false,
        login: vi.fn(),
        register: vi.fn(),
        logout: vi.fn(),
        refreshToken: refreshMock,
      });

      await import("@/lib/api");

      const headers = createHeadersMock();
      const originalRequest = {
        url: "/api/auth/me",
        headers,
        _retry: false,
      } as unknown as InternalAxiosRequestConfig & { _retry?: boolean };

      const error: AxiosError = {
        response: { status: 401 } as AxiosResponse,
        config: originalRequest,
        isAxiosError: true,
        message: "Request failed with status code 401",
        name: "AxiosError",
        toJSON: () => ({}),
      };

      await responseErrorInterceptor(error);
      expect(refreshMock).toHaveBeenCalledOnce();
      expect(mockInstanceFn).toHaveBeenCalledOnce();
    });

    it("does not retry more than once (prevents infinite loop)", async () => {
      const refreshMock = vi.fn().mockResolvedValue("new-token");
      vi.mocked(useAuthStore.getState).mockReturnValue({
        accessToken: "expired-token",
        user: null,
        isAuthenticated: true,
        isLoading: false,
        login: vi.fn(),
        register: vi.fn(),
        logout: vi.fn(),
        refreshToken: refreshMock,
      });

      await import("@/lib/api");

      const headers = createHeadersMock();
      const originalRequest = {
        url: "/api/auth/me",
        headers,
        _retry: true,
      } as unknown as InternalAxiosRequestConfig & { _retry?: boolean };

      const error: AxiosError = {
        response: { status: 401 } as AxiosResponse,
        config: originalRequest,
        isAxiosError: true,
        message: "Request failed with status code 401",
        name: "AxiosError",
        toJSON: () => ({}),
      };

      await expect(responseErrorInterceptor(error)).rejects.toBe(error);
      expect(refreshMock).not.toHaveBeenCalled();
    });

    it("calls logout and rejects when refresh fails", async () => {
      const logoutMock = vi.fn().mockResolvedValue(undefined);
      const refreshMock = vi.fn().mockResolvedValue(null);
      vi.mocked(useAuthStore.getState).mockReturnValue({
        accessToken: "expired-token",
        user: null,
        isAuthenticated: true,
        isLoading: false,
        login: vi.fn(),
        register: vi.fn(),
        logout: logoutMock,
        refreshToken: refreshMock,
      });

      await import("@/lib/api");

      const headers = createHeadersMock();
      const originalRequest = {
        url: "/api/auth/me",
        headers,
        _retry: false,
      } as unknown as InternalAxiosRequestConfig & { _retry?: boolean };

      const error: AxiosError = {
        response: { status: 401 } as AxiosResponse,
        config: originalRequest,
        isAxiosError: true,
        message: "Request failed with status code 401",
        name: "AxiosError",
        toJSON: () => ({}),
      };

      await expect(responseErrorInterceptor(error)).rejects.toBe(error);
      expect(refreshMock).toHaveBeenCalledOnce();
      expect(logoutMock).toHaveBeenCalledOnce();
    });

    it("passes through non-401 errors without refresh", async () => {
      const refreshMock = vi.fn();
      vi.mocked(useAuthStore.getState).mockReturnValue({
        accessToken: "valid-token",
        user: null,
        isAuthenticated: true,
        isLoading: false,
        login: vi.fn(),
        register: vi.fn(),
        logout: vi.fn(),
        refreshToken: refreshMock,
      });

      await import("@/lib/api");

      const error: AxiosError = {
        response: { status: 500 } as AxiosResponse,
        config: { url: "/api/data" } as InternalAxiosRequestConfig,
        isAxiosError: true,
        message: "Server error",
        name: "AxiosError",
        toJSON: () => ({}),
      };

      await expect(responseErrorInterceptor(error)).rejects.toBe(error);
      expect(refreshMock).not.toHaveBeenCalled();
    });
  });
});
