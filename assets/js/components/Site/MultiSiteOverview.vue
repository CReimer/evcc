<template>
	<section class="container px-4 pt-2 pb-5" :aria-label="$t('main.multiSite.title')">
		<div class="row g-3 overview">
			<div v-for="site in sites" :key="site.name" class="col-12 col-md-6 col-xl-4">
				<button
					type="button"
					class="site-card h-100 w-100 p-4 text-start"
					:aria-label="site.title"
					@click="$emit('select', site.name)"
				>
					<div class="d-flex align-items-center justify-content-between mb-4">
						<h2 class="fs-5 mb-0 text-truncate">{{ site.title }}</h2>
						<shopicon-regular-arrowright
							class="arrow ms-3"
						></shopicon-regular-arrowright>
					</div>
					<div class="power-grid">
						<div>
							<div class="label">{{ $t("main.multiSite.pv") }}</div>
							<div class="value solar">{{ fmtW(site.pvPower, POWER_UNIT.AUTO) }}</div>
						</div>
						<div>
							<div class="label">{{ $t("main.multiSite.grid") }}</div>
							<div class="value">{{ fmtW(site.gridPower, POWER_UNIT.AUTO) }}</div>
						</div>
						<div>
							<div class="label">{{ $t("main.multiSite.home") }}</div>
							<div class="value">{{ fmtW(site.homePower, POWER_UNIT.AUTO) }}</div>
						</div>
					</div>
					<div class="loadpoints mt-4 text-muted">
						{{ $t("main.multiSite.loadpoints") }}: {{ site.loadpoints }}
					</div>
				</button>
			</div>
		</div>
	</section>
</template>

<script lang="ts">
import "@h2d2/shopicons/es/regular/arrowright";
import { defineComponent, type PropType } from "vue";
import formatter, { POWER_UNIT } from "@/mixins/formatter";
import type { SiteSummary } from "@/types/evcc";

export default defineComponent({
	name: "MultiSiteOverview",
	mixins: [formatter],
	props: {
		sites: { type: Array as PropType<SiteSummary[]>, default: () => [] },
	},
	emits: ["select"],
	data() {
		return { POWER_UNIT };
	},
});
</script>

<style scoped>
.overview {
	max-width: 72rem;
	margin-inline: auto;
}
.site-card {
	background: var(--evcc-box);
	color: var(--evcc-default-text);
	border: 1px solid var(--bs-border-color-translucent);
	border-radius: 2rem;
	transition: border-color var(--evcc-transition-fast);
}
.site-card:hover,
.site-card:focus-visible {
	border-color: var(--evcc-accent3);
}
.arrow {
	flex: 0 0 auto;
	color: var(--evcc-gray);
}
.power-grid {
	display: grid;
	grid-template-columns: repeat(3, minmax(0, 1fr));
	gap: 1rem;
}
.label,
.loadpoints {
	font-size: 0.8rem;
}
.label {
	color: var(--evcc-gray);
}
.value {
	font-size: 1.25rem;
	font-weight: 600;
	white-space: nowrap;
}
.solar {
	color: var(--evcc-pv);
}
@media (max-width: 380px) {
	.site-card {
		padding: 1.25rem !important;
	}
	.power-grid {
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 1rem;
	}
	.value {
		font-size: 1.05rem;
	}
}
</style>
