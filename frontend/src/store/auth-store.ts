import { create } from "zustand";
import { persist } from "zustand/middleware";
import { apiClient } from "@/lib/api";

export interface User {
  id: string;
  email: string;
  created_at?: string;
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<string | null>;
}

interface AuthResponse {
  access_token: string;
  user: User;
}

interface RefreshResponse {
  access_token: string;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      isAuthenticated: false,
      isLoading: false,

      login: async (email: string, password: string) => {
        set({ isLoading: true });
        try {
          const response = await apiClient.post<AuthResponse>(
            "/api/auth/login",
            { email, password },
          );
          set({
            user: response.data.user,
            accessToken: response.data.access_token,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          set({
            user: null,
            accessToken: null,
            isAuthenticated: false,
            isLoading: false,
          });
          throw error;
        }
      },

      register: async (email: string, password: string) => {
        set({ isLoading: true });
        // Backend contract: register returns 201 with the created user (UserDTO),
        // NOT a token pair — the session is established by an immediate login.
        try {
          await apiClient.post<User>("/api/auth/register", {
            email,
            password,
          });
        } catch (error) {
          set({
            user: null,
            accessToken: null,
            isAuthenticated: false,
            isLoading: false,
          });
          throw error;
        }
        // If this fails, login() already cleared state and re-throws.
        await get().login(email, password);
      },

      logout: async () => {
        try {
          await apiClient.post("/api/auth/logout");
        } catch {
          // Clear state even if API call fails
        } finally {
          set({
            user: null,
            accessToken: null,
            isAuthenticated: false,
          });
        }
      },

      refreshToken: async () => {
        try {
          const response = await apiClient.post<RefreshResponse>(
            "/api/auth/refresh",
          );
          const newToken = response.data.access_token;
          set({ accessToken: newToken });
          return newToken;
        } catch {
          set({
            user: null,
            accessToken: null,
            isAuthenticated: false,
          });
          return null;
        }
      },
    }),
    {
      name: "auth-storage",
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        isAuthenticated: state.isAuthenticated,
      }),
    },
  ),
);
