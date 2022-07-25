<template>
  <DataTable
    :title="`Nodes`"
    :headers="headers"
    :items="nodes"
    :on-click="onClick"
  />
</template>

<script lang="ts">
import { computed, inject, defineComponent } from "@nuxtjs/composition-api";
import DataTable from "./DataTable.vue";
import NodeStoreKey from "../../StoreKey/NodeStoreKey";
import { V1Node } from "@kubernetes/client-node";
import {} from "../../lib/util";

export default defineComponent({
  components: {
    DataTable,
  },
  setup() {
    const nstore = inject(NodeStoreKey);
    if (!nstore) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }

    const nodes = computed(() => nstore.nodes);
    const onClick = (node: V1Node) => {
      nstore.select(node, false);
    };
    const search = "";
    const headers = [
      {
        text: "Name",
        value: "metadata.name",
        sortable: true,
      },
      { text: "CPU", value: "status.capacity.cpu", sortable: true },
      { text: "Memory", value: "status.capacity.memory", sortable: true },
      { text: "Pods", value: "status.capacity.pods", sortable: true },
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
      nodes,
      search,
      headers,
      onClick,
    };
  },
});
</script>
