import { describe, it, expect } from "vitest";
import { readdirSync, statSync } from "fs";
import path from "path";

/**
 * Contract test for the App Router URL map (issue #12).
 *
 * The dashboard was unreachable because `app/(dashboard)/page.tsx` resolves to
 * `/` — route groups in parentheses add no URL segment — and collided with
 * `app/page.tsx`. Component-render tests could never see that: they mount the
 * component directly and never ask the router whether the URL exists.
 *
 * So instead of mocking the router, we derive the real route map from the
 * filesystem using App Router's rules and assert against it.
 */

const APP_DIR = path.resolve(__dirname, "../app");

function collectPages(dir: string): string[] {
  const out: string[] = [];
  for (const entry of readdirSync(dir)) {
    const full = path.join(dir, entry);
    if (statSync(full).isDirectory()) {
      out.push(...collectPages(full));
    } else if (entry === "page.tsx" || entry === "page.ts") {
      out.push(full);
    }
  }
  return out;
}

/** app/(dashboard)/dashboard/page.tsx -> /dashboard ; app/page.tsx -> / */
function routeFor(pageFile: string): string {
  const rel = path
    .relative(APP_DIR, path.dirname(pageFile))
    .split(path.sep)
    .filter((seg) => !seg.startsWith("(") && seg !== "");
  return rel.length === 0 ? "/" : "/" + rel.join("/");
}

const routes = collectPages(APP_DIR).map(routeFor);

describe("route map", () => {
  it("has no two pages resolving to the same URL", () => {
    const duplicates = [
      ...new Set(routes.filter((r) => routes.indexOf(r) !== routes.lastIndexOf(r))),
    ];
    expect(
      duplicates,
      `pages colliding on the same URL: ${duplicates.join(", ")}`,
    ).toEqual([]);
  });

  it("serves the dashboard at /dashboard, the URL login and register navigate to", () => {
    expect(routes).toContain("/dashboard");
  });

  it("resolves every URL that the app navigates to at runtime", () => {
    // login/register push /dashboard; the root page redirects to /dashboard;
    // the proxy redirects to /login. If any target is not a real route, the
    // user lands on a 404 with a green test suite.
    const targets = ["/dashboard", "/login"];
    const missing = targets.filter((t) => !routes.includes(t));
    expect(missing, `navigated URLs with no page: ${missing.join(", ")}`).toEqual([]);
  });
});
