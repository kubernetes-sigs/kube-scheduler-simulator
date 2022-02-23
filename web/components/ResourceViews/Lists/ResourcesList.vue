<template>
  <div>
    <NodeList />
    <UnscheduledPodList />
    <PVList />
    <PVCList />
    <StorageClassList />
    <PriorityClassList />
  </div>
</template>

<script lang="ts">
import { defineComponent } from "@nuxtjs/composition-api";
import NodeList from "./NodeList.vue";
import UnscheduledPodList from "./UnscheduledPodList.vue";
import PVList from "./PVList.vue";
import PVCList from "./PVCList.vue";
import StorageClassList from "./StorageClassList.vue";
import PriorityClassList from "./PriorityClassList.vue";

export default defineComponent({
  components: {
    NodeList,
    UnscheduledPodList,
    PVList,
    PVCList,
    StorageClassList,
    PriorityClassList,
  },
});
</script>
