import { shallowMount } from "@vue/test-utils";
import { describe, expect, test } from "vite-plus/test";
import SiteSelector from "./SiteSelector.vue";

describe("SiteSelector", () => {
  test("renders site titles and emits the selected site", async () => {
    const wrapper = shallowMount(SiteSelector, {
      global: { mocks: { $t: (key: string) => key } },
      props: {
        siteNames: ["home", "office"],
        activeSite: "home",
        siteTitles: { home: "Home", office: "Office" },
      },
    });

    const select = wrapper.get("select");
    expect(select.element.value).toBe("home");
    expect(select.text()).toContain("Office");
    await select.setValue("office");
    expect(wrapper.emitted("select")).toEqual([["office"]]);
  });

  test("can limit selection to individual sites", () => {
    const wrapper = shallowMount(SiteSelector, {
      global: { mocks: { $t: (key: string) => key } },
      props: {
        siteNames: ["home", "office"],
        activeSite: "home",
        includeAggregate: false,
      },
    });

    expect(wrapper.find('option[value="__all__"]').exists()).toBe(false);
    expect(wrapper.findAll("option")).toHaveLength(2);
  });
});
