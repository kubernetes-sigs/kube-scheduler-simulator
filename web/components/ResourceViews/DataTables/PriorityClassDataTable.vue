<template>
  <v-row>
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1">
          PriorityClasses <v-spacer></v-spacer>
          <v-text-field
            v-model="search"
            append-icon="mdi-magnify"
            label="Search"
            single-line
            hide-details
          ></v-text-field
        ></v-card-title>
        <v-data-table
          :headers="headers"
          :items="priorityclasses"
          :items-per-page="5"
          :search="search"
          multi-sort
          @click:row="onClick"
        ></v-data-table>
      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { V1PriorityClass } from "@kubernetes/client-node";
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import {} from "../../lib/util";
import PriorityClassStoreKey from "../../StoreKey/PriorityClassStoreKey";

export default defineComponent({
  setup() {
    const store = inject(PriorityClassStoreKey);
    if (!store) {
      throw new Error(`${PriorityClassStoreKey} is not provided`);
    }

    const getPriorityClassList = async () => {
      await store.fetchlist();
    };
    const onClick = (priorityclass: V1PriorityClass) => {
      store.select(priorityclass, false);
    };
    onMounted(getPriorityClassList);
    const priorityclasses = computed(() => store.priorityclasses);
    const search = "";
    const headers = [
      {
        text: "Name",
        value: "metadata.name",
        sortable: true,
      },
      { text: "Namespace", value: "metadata.namespace", sortable: true },
      { text: "Value", value: "value", sortable: true },
      { text: "Global-Default", value: "globalDefault", sortable: true },
      { text: "Preemption-Policy", value: "preemptionPolicy", sortable: true },
      {
        text: "Creation-Time",
        value: "metadata.creationTimestamp",
        sortable: true,
      },
      {
        text: "Update-Time",
        value: "metadata.managedFields[0].time",
        sortable: true,
      },
    ];
    return {
      priorityclasses,
      onClick,
      search,
      headers,
    };
  },
});
</script>
