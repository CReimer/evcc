<template>
	<label class="site-selector d-flex align-items-center">
		<span class="visually-hidden">{{ $t("main.siteSelector.label") }}</span>
		<select
			class="form-select border-0 bg-transparent fw-bold p-0 pe-5"
			:class="{ compact }"
			:value="activeSite"
			:aria-label="$t('main.siteSelector.label')"
			@change="selectSite"
		>
			<option v-if="includeAggregate" value="__all__">
				{{ $t("main.siteSelector.all") }}
			</option>
			<option v-for="site in siteNames" :key="site" :value="site">
				{{ siteTitle(site) }}
			</option>
		</select>
	</label>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

export default defineComponent({
	name: "SiteSelector",
	props: {
		siteNames: { type: Array as PropType<string[]>, required: true },
		activeSite: { type: String, required: true },
		siteTitles: { type: Object as PropType<Record<string, string>>, default: () => ({}) },
		includeAggregate: { type: Boolean, default: true },
		compact: Boolean,
	},
	emits: ["select"],
	methods: {
		siteTitle(name: string) {
			return this.siteTitles[name] || name;
		},
		selectSite(event: Event) {
			this.$emit("select", (event.target as HTMLSelectElement).value);
		},
	},
});
</script>

<style scoped>
.site-selector {
	min-width: 0;
}
.site-selector select {
	max-width: min(15rem, 58vw);
	font-size: 1.5rem;
	color: var(--evcc-default-text);
	box-shadow: none;
}
.site-selector select.compact {
	font-size: 1rem;
	font-weight: 500 !important;
}
</style>
