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

<script lng="ts">
import { defineComponent } from "@nuxtjs/composition-api";
import NodeList from "~/components/ResourceViews/List/NodeList.vue";
import UnscheduledPodList from "~/components/ResourceViews/List/UnscheduledPodList.vue";
import PVList from "~/components/ResourceViews/List/PVList.vue";
import PVCList from "~/components/ResourceViews/List/PVCList.vue";
import StorageClassList from "~/components/ResourceViews/List/StorageClassList.vue";
import PriorityClassList from "~/components/ResourceViews/List/PriorityClassList.vue";

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
