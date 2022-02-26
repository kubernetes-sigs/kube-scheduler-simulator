<template>
  <v-sheet class="transparent">
    <SchedulerConfigurationEditButton />
    <ExportButton />
    <ImportButton />
  </v-sheet>
</template>

<script lang="ts">
import { defineComponent } from "@nuxtjs/composition-api";
import SchedulerConfigurationEditButton from "./SchedulerConfigurationEditButton.vue";
import ExportButton from "./ExportButton.vue";
import ImportButton from "./ImportButton.vue";

export default defineComponent({
  components: {
    ImportButton,
    ExportButton,
    SchedulerConfigurationEditButton,
  },
});
</script>
