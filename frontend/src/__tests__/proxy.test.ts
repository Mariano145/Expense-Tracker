import { describe, it, expect } from "vitest";
import { NextRequest } from "next/server";
import { proxy } from "../proxy";

function createRequest(path: string, cookies?: Record<string, string>) {
  const url = `http://localhost:3000${path}`;
  const request = new NextRequest(url);
  if (cookies) {
    for (const [key, value] of Object.entries(cookies)) {
      request.cookies.set(key, value);
    }
  }
  return request;
}

describe("proxy", () => {
  describe("protected routes", () => {
    it("redirects to /login when accessing /dashboard without auth cookie", () => {
      const request = createRequest("/dashboard");
      const response = proxy(request);

      expect(response.status).toBe(307);
      expect(response.headers.get("location")).toContain("/login");
    });

    it("redirects to /login when accessing /dashboard sub-routes without auth cookie", () => {
      const request = createRequest("/dashboard/settings");
      const response = proxy(request);

      expect(response.status).toBe(307);
      expect(response.headers.get("location")).toContain("/login");
    });

    it("allows access to /dashboard when refresh_token cookie is present", () => {
      const request = createRequest("/dashboard", {
        refresh_token: "valid-token",
      });
      const response = proxy(request);

      expect(response.status).toBe(200);
    });
  });

  describe("auth routes", () => {
    it("redirects to /dashboard when accessing /login with auth cookie", () => {
      const request = createRequest("/login", {
        refresh_token: "valid-token",
      });
      const response = proxy(request);

      expect(response.status).toBe(307);
      expect(response.headers.get("location")).toContain("/dashboard");
    });

    it("redirects to /dashboard when accessing /register with auth cookie", () => {
      const request = createRequest("/register", {
        refresh_token: "valid-token",
      });
      const response = proxy(request);

      expect(response.status).toBe(307);
      expect(response.headers.get("location")).toContain("/dashboard");
    });

    it("allows access to /login without auth cookie", () => {
      const request = createRequest("/login");
      const response = proxy(request);

      expect(response.status).toBe(200);
    });

    it("allows access to /register without auth cookie", () => {
      const request = createRequest("/register");
      const response = proxy(request);

      expect(response.status).toBe(200);
    });
  });

  describe("other routes", () => {
    it("allows access to public routes", () => {
      const request = createRequest("/");
      const response = proxy(request);

      expect(response.status).toBe(200);
    });
  });
});
