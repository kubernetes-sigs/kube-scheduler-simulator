<template>
  <v-sheet class="transparent">
    <SchedulerConfigurationEditButton />
    <ExportButton />
    <ImportButton />
    <ResetButton />
  </v-sheet>
</template>

<script lang="ts">
import { defineComponent } from "@nuxtjs/composition-api";
import SchedulerConfigurationEditButton from "./SchedulerConfigurationEditButton.vue";
import ExportButton from "./ExportButton.vue";
import ImportButton from "./ImportButton.vue";
import ResetButton from "./ResetButton.vue";

export default defineComponent({
  components: {
    ImportButton,
    ExportButton,
    SchedulerConfigurationEditButton,
    ResetButton,
  },
});
</script>
