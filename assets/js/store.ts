import { computed, reactive, ref, watch } from "vue";
import type { State } from "./types/evcc";
import { convertToUiLoadpoints } from "./uiLoadpoints";
import { useDebouncedComputed } from "./utils/useDebouncedComputed";
import { expandForecast } from "./utils/forecast";
import settings from "./settings";
import { aggregateState, stateForSite, summarizeSites } from "./utils/sites";
import { setActiveSite } from "./utils/siteApi";

function setProperty(obj: object, props: string[], value: any) {
  const prop = props.shift();
  // @ts-expect-error no-explicit-any
  if (!obj[prop]) {
    // @ts-expect-error no-explicit-any
    obj[prop] = props.length > 0 && /^\d+$/.test(props[0]) ? [] : {};
  }

  if (!props.length) {
    if (value && typeof value === "object" && !Array.isArray(value)) {
      // @ts-expect-error no-explicit-any
      obj[prop] = { ...obj[prop], ...value };
    } else {
      // @ts-expect-error no-explicit-any
      obj[prop] = value;
    }
    return;
  }

  // @ts-expect-error no-explicit-any
  setProperty(obj[prop], props, value);
}

const initialState: State = {
  offline: false,
  loadpoints: [],
  vehicles: {},
  forecast: {},
};

const state = reactive(initialState);
const aggregateSite = "__all__";
const selectedSite = ref(localStorage.getItem("selectedSite") || "");

const siteNames = computed(() => state.siteNames || []);

const activeSiteName = computed(() => {
  const names = siteNames.value;
  if (names.length < 2) return names[0] || "";
  if (selectedSite.value === aggregateSite || names.includes(selectedSite.value)) {
    return selectedSite.value;
  }
  return names[0];
});

const siteSummaries = computed(() => summarizeSites(state, siteNames.value));

const activeState = computed<State>(() => {
  const name = activeSiteName.value;
  if (name === aggregateSite) {
    return aggregateState(state, siteSummaries.value);
  }
  if (!name) return state;
  return { ...state, ...stateForSite(state, name), aggregateMode: false };
});

watch(
  activeSiteName,
  (name) => {
    const apiSite = name && name !== aggregateSite && siteNames.value.length > 1 ? name : null;
    setActiveSite(apiSite);
    selectedSite.value = name;
    if (name) localStorage.setItem("selectedSite", name);
  },
  { immediate: true }
);

// create derived loadpoints array with ui specific fields (defaults, browser settings, ...); debounce for better performance
const uiLoadpoints = useDebouncedComputed(
  () => convertToUiLoadpoints(activeState.value.loadpoints, activeState.value.vehicles),
  () => [activeState.value.loadpoints, activeState.value.vehicles, settings.loadpoints],
  50
);

// derived forecast with slots expanded to objects with unix milliseconds; lazy,
// only computed while a forecast consumer is mounted
const uiForecast = computed(() => expandForecast(activeState.value.forecast));

export interface Store {
  state: State; // raw state from websocket
  activeState: typeof activeState;
  selectedSite: typeof selectedSite;
  activeSiteName: typeof activeSiteName;
  siteNames: typeof siteNames;
  siteSummaries: typeof siteSummaries;
  selectSite(name: string): void;
  uiLoadpoints: typeof uiLoadpoints;
  uiForecast: typeof uiForecast;
  offline(value: boolean): void;
  update(msg: any): void;
  reset(): void;
}

const store: Store = {
  state,
  activeState,
  selectedSite,
  activeSiteName,
  siteNames,
  siteSummaries,
  selectSite(name: string) {
    selectedSite.value = name;
  },
  uiLoadpoints,
  uiForecast,
  offline(value: boolean) {
    state.offline = value;
  },
  update(msg) {
    Object.keys(msg).forEach(function (k) {
      if (k === "log") {
        window.app.raise(msg[k]);
      } else {
        setProperty(state, k.split("."), msg[k]);
      }
    });
  },
  reset() {
    console.log("resetting state");
    // reset to initial state
    Object.keys(initialState).forEach(function (k) {
      if (k === "offline") return;

      // @ts-expect-error no-explicit-any
      if (Array.isArray(initialState[k])) {
        // @ts-expect-error no-explicit-any
        state[k] = [];
      } else {
        // @ts-expect-error no-explicit-any
        state[k] = undefined;
      }
    });
  },
};

export default store;
