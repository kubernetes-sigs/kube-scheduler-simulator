<template>
  <div>
    <NodeDataTable />
    <UnscheduledPodDataTable />
    <PVDataTable />
    <PVCDataTable />
    <StorageClassDataTable />
    <PriorityClassDataTable />
  </div>
</template>

<script lng="ts">
import { defineComponent } from "@nuxtjs/composition-api";
import NodeDataTable from "./NodeDataTable.vue";
import UnscheduledPodDataTable from "./UnscheduledPodDataTable.vue";
import PVDataTable from "./PVDataTable.vue";
import PVCDataTable from "./PVCDataTable.vue";
import StorageClassDataTable from "./StorageClassDataTable.vue";
import PriorityClassDataTable from "./PriorityClassDataTable.vue";

export default defineComponent({
  components: {
    NodeDataTable,
    UnscheduledPodDataTable,
    PVDataTable,
    PVCDataTable,
    StorageClassDataTable,
    PriorityClassDataTable,
  },
});
</script>
