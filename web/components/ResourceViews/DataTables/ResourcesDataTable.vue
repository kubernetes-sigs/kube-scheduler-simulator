<template>
  <div>
    <PodDataTable />
    <NodeDataTable />
    <PVDataTable />
    <PVCDataTable />
    <StorageClassDataTable />
    <PriorityClassDataTable />
    <NamespaceDataTable />
  </div>
</template>

<script lang="ts">
import { defineComponent } from "@nuxtjs/composition-api";
import PodDataTable from "./PodDataTable.vue";
import NodeDataTable from "./NodeDataTable.vue";
import PVDataTable from "./PVDataTable.vue";
import PVCDataTable from "./PVCDataTable.vue";
import StorageClassDataTable from "./StorageClassDataTable.vue";
import PriorityClassDataTable from "./PriorityClassDataTable.vue";
import NamespaceDataTable from "./NamespaceDataTable.vue";

export default defineComponent({
  components: {
    PodDataTable,
    NodeDataTable,
    PVDataTable,
    PVCDataTable,
    StorageClassDataTable,
    PriorityClassDataTable,
    NamespaceDataTable,
  },
});
</script>
