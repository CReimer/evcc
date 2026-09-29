import { describe, expect, test } from "vite-plus/test";
import { siteApiUrl } from "./siteApi";

describe("siteApiUrl", () => {
  test("namespaces site control and history APIs", () => {
    expect(siteApiUrl("loadpoints/1/mode/pv", "office")).toBe("sites/office/loadpoints/1/mode/pv");
    expect(siteApiUrl("/sessions?year=2026", "office east")).toBe(
      "sites/office%20east/sessions?year=2026"
    );
  });

  test("leaves global APIs and aggregate mode unchanged", () => {
    expect(siteApiUrl("health", "office")).toBe("health");
    expect(siteApiUrl("loadpoints/1/mode/pv", null)).toBe("loadpoints/1/mode/pv");
  });

  test("adds the site to privileged history deletion", () => {
    expect(siteApiUrl("db/metrics?group=pv", "home_1")).toBe("db/metrics?group=pv&site=home_1");
  });
});
