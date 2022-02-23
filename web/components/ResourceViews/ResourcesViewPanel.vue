<template>
  <div>
    <ViewSelectRadioButton
      v-model="selectedViewType"
      :radio-labels="viewTypeLabels"
    />
    <div v-if="selectedViewType === 0">
      <ResourcesList />
    </div>
    <div v-else-if="selectedViewType === 1">
      <!-- TODO: -->
    </div>
  </div>
</template>

<script lng="ts">
import { defineComponent, ref } from "@nuxtjs/composition-api";
import ResourcesList from "./Lists/ResourcesList.vue";
import ViewSelectRadioButton from "./ViewSeelectRadioButton.vue";

export default defineComponent({
  components: {
    ResourcesList,
    ViewSelectRadioButton,
  },
  setup() {
    const viewTypeLabels = ["List", "Table"];
    const selectedViewType = ref(0);
    return {
      viewTypeLabels,
      selectedViewType,
    };
  },
});
</script>
