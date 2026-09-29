<template>
	<header
		class="d-flex justify-content-between align-items-center py-3 py-md-4"
		data-testid="header"
	>
		<div class="title position-relative flex-grow-1 mb-1">
			<Transition name="fade-swap-left">
				<SiteSelector
					v-if="!parentTitle && siteNames.length > 1 && siteSelectorAsTitle"
					:site-names="siteNames"
					:active-site="activeSite"
					:site-titles="siteTitles"
					:include-aggregate="includeAggregate"
					@select="$emit('site-select', $event)"
				/>
				<div v-else-if="!parentTitle">
					<h1 class="mb-0 pt-1 text-truncate">{{ title }}</h1>
					<SiteSelector
						v-if="siteNames.length > 1"
						compact
						:site-names="siteNames"
						:active-site="activeSite"
						:site-titles="siteTitles"
						:include-aggregate="includeAggregate"
						@select="$emit('site-select', $event)"
					/>
				</div>
			</Transition>
			<Transition name="fade-swap-right">
				<div v-if="parentTitle">
					<div class="d-flex align-items-center text-nowrap">
						<button
							type="button"
							class="btn btn-link back-button d-flex align-items-center p-0 border-0 me-2"
							:aria-label="$t('general.back')"
							@click="$emit('back')"
						>
							<shopicon-bold-arrowback></shopicon-bold-arrowback>
						</button>
						<h1 class="mb-0 pt-1 text-truncate">{{ title }}</h1>
					</div>
					<SiteSelector
						v-if="siteNames.length > 1 && !siteSelectorAsTitle"
						compact
						:site-names="siteNames"
						:active-site="activeSite"
						:site-titles="siteTitles"
						:include-aggregate="includeAggregate"
						@select="$emit('site-select', $event)"
					/>
				</div>
			</Transition>
		</div>
		<TopNavigationArea
			ref="navigationArea"
			:notifications="notifications"
			:show-savings="showSavings"
		/>
	</header>
</template>

<script lang="ts">
import "@h2d2/shopicons/es/bold/arrowback";
import TopNavigationArea from "./TopNavigationArea.vue";
import SiteSelector from "./SiteSelector.vue";
import { defineComponent, type PropType } from "vue";
import type { Notification } from "@/types/evcc";

export default defineComponent({
	name: "TopHeader",
	components: {
		TopNavigationArea,
		SiteSelector,
	},
	props: {
		title: String,
		// set on drill-down pages: shows a back button next to the title
		parentTitle: String,
		notifications: { type: Array as PropType<Notification[]>, default: () => [] },
		siteNames: { type: Array as PropType<string[]>, default: () => [] },
		activeSite: { type: String, default: "" },
		siteTitles: { type: Object as PropType<Record<string, string>>, default: () => ({}) },
		showSavings: { type: Boolean, default: true },
		siteSelectorAsTitle: { type: Boolean, default: true },
		includeAggregate: { type: Boolean, default: true },
	},
	emits: ["back", "site-select"],
	methods: {
		requestAuthProvider(providerId: string) {
			const navigationArea = this.$refs["navigationArea"] as
				| InstanceType<typeof TopNavigationArea>
				| undefined;
			navigationArea?.requestAuthProvider(providerId);
		},
	},
});
</script>

<style scoped>
/* allow the flex item to shrink so the titles can truncate */
.title {
	min-width: 0;
}
.back-button {
	color: var(--evcc-default-text);
}
</style>
