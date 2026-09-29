<template>
	<JsonModal
		name="title"
		:title="$t('config.title.title')"
		:endpoint="endpoint"
		save-method="put"
		:transform-read-values="transformReadValues"
		disable-remove
		@changed="$emit('changed')"
	>
		<template #default="{ values }">
			<FormRow
				id="siteTitle"
				:label="$t('config.title.label')"
				:help="$t('config.title.description')"
			>
				<input id="siteTitle" v-model="values.title" class="form-control" />
			</FormRow>
		</template>
	</JsonModal>
</template>

<script>
import JsonModal from "./JsonModal.vue";
import FormRow from "./FormRow.vue";

export default {
	name: "TitleModal",
	components: { FormRow, JsonModal },
	props: {
		endpoint: { type: String, default: "/config/site" },
		value: { type: String, default: "" },
	},
	emits: ["changed"],
	methods: {
		transformReadValues() {
			return { title: this.value };
		},
	},
};
</script>
