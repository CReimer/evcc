import { shallowMount } from "@vue/test-utils";
import { describe, expect, test } from "vite-plus/test";
import MultiSiteOverview from "./MultiSiteOverview.vue";

describe("MultiSiteOverview", () => {
  test("opens a site from its summary card", async () => {
    const wrapper = shallowMount(MultiSiteOverview, {
      global: { mocks: { $t: (key: string) => key, $i18n: { locale: "en" } } },
      props: {
        sites: [
          {
            name: "home",
            title: "Home",
            gridPower: 200,
            pvPower: 1200,
            homePower: 1000,
            loadpoints: 1,
          },
        ],
      },
    });

    await wrapper.get("button").trigger("click");
    expect(wrapper.emitted("select")).toEqual([["home"]]);
  });
});
