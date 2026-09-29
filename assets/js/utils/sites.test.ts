import { describe, expect, test } from "vite-plus/test";
import type { State } from "@/types/evcc";
import { aggregateState, stateForSite, summarizeSites } from "./sites";

const state = {
  offline: false,
  loadpoints: [{ id: 1 }],
  vehicles: {},
  forecast: {},
  siteTitle: "Home",
  grid: { power: 1200 },
  pvPower: 800,
  homePower: 400,
  sites: {
    home: {
      siteTitle: "Home",
      grid: { power: 1200 },
      pvPower: 800,
      homePower: 400,
      loadpoints: [{ id: 1 }],
    },
    office: {
      siteTitle: "Office",
      grid: { power: -300 },
      pvPower: 1500,
      homePower: 900,
      loadpoints: [{ id: 1 }, { id: 2 }],
    },
  },
} as unknown as State;

describe("multi-site state", () => {
  test("resolves primary and additional sites", () => {
    expect(stateForSite(state, "home").siteTitle).toBe("Home");
    expect(stateForSite(state, "office").siteTitle).toBe("Office");
    expect(stateForSite(state, "office").vehicles).toBe(state.vehicles);
  });

  test("does not inherit primary devices for an empty site", () => {
    const emptySite = stateForSite({ ...state, sites: { empty: { siteTitle: "Empty" } } }, "empty");

    expect(emptySite.loadpoints).toEqual([]);
    expect(emptySite.vehicles).toBe(state.vehicles);
  });

  test("builds a read-only aggregate without loadpoints or vehicles", () => {
    const summaries = summarizeSites(state, ["home", "office"]);
    const aggregate = aggregateState(state, summaries);

    expect(summaries.map(({ name, loadpoints }) => ({ name, loadpoints }))).toEqual([
      { name: "home", loadpoints: 1 },
      { name: "office", loadpoints: 2 },
    ]);
    expect(aggregate.grid?.power).toBe(900);
    expect(aggregate.pvPower).toBe(2300);
    expect(aggregate.homePower).toBe(1300);
    expect(aggregate.loadpoints).toEqual([]);
    expect(aggregate.vehicles).toEqual({});
    expect(aggregate.statistics).toBeUndefined();
    expect(aggregate.forecast).toEqual({});
  });
});
