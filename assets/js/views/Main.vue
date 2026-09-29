<template>
	<Site
		:notifications="notifications"
		v-bind="state"
		:forecast="uiForecast"
		:selected-loadpoint-index="selectedLoadpointIndex"
		:site-names="siteNames"
		:active-site="activeSite"
		:site-titles="siteTitles"
		:ui-loadpoints="uiLoadpoints"
		@site-select="selectSite"
	/>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import Site from "../components/Site/Site.vue";
import store from "../store";
import type { Notification } from "@/types/evcc";

export default defineComponent({
	name: "Main",
	components: { Site },
	props: {
		notifications: Array as PropType<Notification[]>,
		selectedLoadpointIndex: Number,
	},
	computed: {
		state() {
			return store.activeState.value;
		},
		uiForecast() {
			return store.uiForecast.value;
		},
		uiLoadpoints() {
			return store.uiLoadpoints.value;
		},
		siteNames() {
			return store.siteNames.value;
		},
		activeSite() {
			return store.activeSiteName.value;
		},
		siteTitles() {
			return Object.fromEntries(
				store.siteSummaries.value.map(({ name, title }) => [name, title])
			);
		},
	},
	methods: {
		selectSite(name: string) {
			store.selectSite(name);
		},
	},
	head() {
		const title = store.activeState.value.siteTitle;
		if (title) {
			return { title };
		}
		// no custom title
		return { title: "evcc", titleTemplate: null };
	},
});
</script>
