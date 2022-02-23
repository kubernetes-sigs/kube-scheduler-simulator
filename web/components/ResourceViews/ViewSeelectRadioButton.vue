<template>
  <v-container fluid>
    <v-radio-group v-model="viewTypeComputed" row>
      <v-radio
        v-for="(_, i) in radioLabels.length"
        :key="i"
        :label="radioLabels[i]"
        :value="i"
      ></v-radio>
    </v-radio-group>
  </v-container>
</template>

<script lng="ts">
import { defineComponent, computed } from "@nuxtjs/composition-api";

export default defineComponent({
  props: {
    radioLabels: {
      type: Array,
      default: () => [],
    },
    selectedViewType: {
      type: Number,
      default: 0,
    },
  },
  emits: ["input"],
  setup(props, { emit }) {
    const viewTypeComputed = computed({
      get: () => props.selectedViewType,
      set: (newValue) => emit("input", newValue),
    });
    return {
      viewTypeComputed,
    };
  },
});
</script>
