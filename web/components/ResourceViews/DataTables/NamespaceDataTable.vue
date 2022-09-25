<template>
  <DataTable
    :title="`Namespaces`"
    :headers="headers"
    :items="namespaces"
    :on-click="onClick"
  />
</template>

<script lang="ts">
import { V1Namespace } from "@kubernetes/client-node";
import { computed, inject, defineComponent } from "@nuxtjs/composition-api";
import DataTable from "./DataTable.vue";
import NamespaceStoreKey from "../../StoreKey/NamespaceStoreKey";

export default defineComponent({
  components: {
    DataTable,
  },
  setup() {
    const store = inject(NamespaceStoreKey);
    if (!store) {
      throw new Error(`${NamespaceStoreKey.description} is not provided`)
    }
    const onClick = (ns: V1Namespace) => {
      store.select(ns, false)
    };

    const namespaces = computed(() => store.namespaces);
    const search = "";
    const headers = [
      {
        text: "Name",
        value: "metadata.name",
        sortable: true,
      },
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
      namespaces,
      onClick,
      search,
      headers,
    };
  },
});
</script>
