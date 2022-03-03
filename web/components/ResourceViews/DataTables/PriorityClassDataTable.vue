<template>
  <DataTable
    :title="`PriorityClasses`"
    :headers="headers"
    :items="priorityclasses"
    :on-click="onClick"
  />
</template>

<script lang="ts">
import { V1PriorityClass } from "@kubernetes/client-node";
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import DataTable from "./DataTable.vue";
import {} from "../../lib/util";
import PriorityClassStoreKey from "../../StoreKey/PriorityClassStoreKey";

export default defineComponent({
  components: {
    DataTable,
  },
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
      { text: "Value", value: "value", sortable: true },
      { text: "GlobalDefault", value: "globalDefault", sortable: true },
      { text: "PreemptionPolicy", value: "preemptionPolicy", sortable: true },
      {
        text: "CreationTime",
        value: "metadata.creationTimestamp",
        sortable: true,
      },
      {
        text: "UpdateTime",
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
