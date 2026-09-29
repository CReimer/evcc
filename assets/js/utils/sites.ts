import type { ConfigLoadpoint, SiteSummary, State } from "@/types/evcc";

export function stateForSite(state: State, name: string): Partial<State> {
  return { loadpoints: [], ...state.sites?.[name], vehicles: state.vehicles };
}

export function loadpointsForSite(
  loadpoints: ConfigLoadpoint[],
  siteNames: string[],
  activeSite: string
) {
  const primary = siteNames[0];
  return loadpoints.filter((loadpoint) => (loadpoint.site || primary) === activeSite);
}

export function summarizeSites(state: State, names: string[]): SiteSummary[] {
  return names.map((name) => {
    const site = stateForSite(state, name);
    return {
      name,
      title: site.siteTitle || name,
      gridPower: site.grid?.power || 0,
      pvPower: site.pvPower || 0,
      homePower: site.homePower || 0,
      loadpoints: site.loadpoints?.length || 0,
    };
  });
}

export function aggregateState(state: State, summaries: SiteSummary[]): State {
  const sum = (key: "gridPower" | "pvPower" | "homePower") =>
    summaries.reduce((total, site) => total + site[key], 0);
  return {
    ...state,
    aggregateMode: true,
    siteTitle: "",
    grid: { power: sum("gridPower") },
    gridConfigured: summaries.length > 0,
    pvPower: sum("pvPower"),
    homePower: sum("homePower"),
    loadpoints: [],
    vehicles: {},
    battery: undefined,
    forecast: {},
    statistics: undefined,
    tariffGrid: undefined,
    tariffFeedIn: undefined,
    tariffCo2: undefined,
    tariffPriceHome: undefined,
    tariffCo2Home: undefined,
    tariffPriceLoadpoints: undefined,
    tariffCo2Loadpoints: undefined,
    siteSummaries: summaries,
  };
}
