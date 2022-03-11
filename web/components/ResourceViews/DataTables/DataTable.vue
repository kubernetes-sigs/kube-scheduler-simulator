<template>
  <v-row>
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1">
          <v-row
            ><v-col> {{ title }} <v-spacer></v-spacer> </v-col
            ><v-col>
              <v-text-field
                v-model="search"
                append-icon="mdi-magnify"
                label="Search"
                single-line
                hide-details
              ></v-text-field></v-col></v-row></v-card-title
        ><v-data-table
          :headers="headers"
          :items="items"
          :items-per-page="5"
          :search="search"
          multi-sort
          class="row-pointer"
          @click:row="onClick"
        ></v-data-table>
      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { defineComponent } from "@nuxtjs/composition-api";

export default defineComponent({
  props: {
    headers: {
      type: Array,
      default: () => [],
    },
    title: {
      type: String,
      default: "",
    },
    items: {
      type: Array,
      default: () => [],
    },
    onClick: {
      type: Function,
      default: () => {},
    },
  },
  setup() {
    const search = "";

    return {
      search,
    };
  },
});
</script>

<style>
.row-pointer > .v-data-table__wrapper > table > tbody > tr:hover {
  cursor: pointer;
}
</style>
