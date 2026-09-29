import { shallowMount } from "@vue/test-utils";
import { afterEach, describe, expect, test } from "vite-plus/test";
import { nextTick } from "vue";
import Main from "./Main.vue";
import Site from "../components/Site/Site.vue";
import store from "../store";

describe("multi-site view", () => {
  afterEach(() => {
    store.selectSite("");
    store.state.siteNames = [];
    store.state.sites = {};
    store.state.loadpoints = [];
    store.state.vehicles = {};
  });

  test("renders the selected site state", async () => {
    store.update({
      siteNames: ["mt8", "mt4"],
      sites: {
        mt8: {
          siteTitle: "MT8",
          loadpoints: [{ id: 1 }],
          vehicles: {},
        },
        mt4: {
          siteTitle: "MT4",
        },
      },
    });
    store.selectSite("mt4");

    const wrapper = shallowMount(Main);
    await nextTick();

    const site = wrapper.findComponent(Site);
    expect(site.props("siteTitle")).toBe("MT4");
    expect(site.props("vehicles")).toEqual({});
    expect(store.activeState.value.loadpoints).toEqual([]);
    expect(site.props("siteNames")).toEqual(["mt8", "mt4"]);
  });

  test("uses the nested state for the primary site", () => {
    store.update({
      siteNames: ["mt8", "mt4"],
      loadpoints: [{ id: 99 }],
      vehicles: { shared: { title: "Shared" } },
      sites: {
        mt8: {
          siteTitle: "MT8",
          loadpoints: [{ id: 1 }],
        },
      },
    });
    store.selectSite("mt8");

    expect(store.activeState.value.loadpoints).toEqual([{ id: 1 }]);
    expect(store.activeState.value.vehicles).toEqual({ shared: { title: "Shared" } });
  });

  test("builds nested loadpoint arrays from websocket paths", () => {
    store.update({
      siteNames: ["mt8", "mt4"],
      "sites.mt8.loadpoints.0.id": 1,
      "sites.mt8.loadpoints.0.title": "Parking",
    });

    expect(store.siteSummaries.value[0].loadpoints).toBe(1);
    expect(store.state.sites?.["mt8"]?.loadpoints).toEqual([{ id: 1, title: "Parking" }]);
  });
});
