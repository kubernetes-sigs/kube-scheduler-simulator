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
import SnackBarStoreKey from "../../StoreKey/SnackBarStoreKey";

export default defineComponent({
  components: {
    DataTable,
  },
  setup() {
    const store = inject(PriorityClassStoreKey);
    if (!store) {
      throw new Error(`${PriorityClassStoreKey} is not provided`);
    }

    const snackbarstore = inject(SnackBarStoreKey);
    if (!snackbarstore) {
      throw new Error(`${SnackBarStoreKey} is not provided`);
    }

    const setServerErrorMessage = (error: string) => {
      snackbarstore.setServerErrorMessage(error);
    };

    const getPriorityClassList = async () => {
      await store.fetchlist().catch((e) => setServerErrorMessage(e));
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
