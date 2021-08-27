<template>
  <v-sheet>
    <v-card-title>Resource Definition</v-card-title>
    <v-treeview dense :items="items"></v-treeview>
  </v-sheet>
</template>

<script lang="ts">
import { defineComponent } from "@nuxtjs/composition-api";

export default defineComponent({
  props: {
    items: {
      type: Array,
      required: true,
    },
  },
});
</script>
