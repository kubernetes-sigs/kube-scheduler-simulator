<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide, useContext } from "@nuxtjs/composition-api";
import priorityClassStore from "../../store/priorityclass";
import PriorityClassStoreKey from "../StoreKey/PriorityClassStoreKey";

export default defineComponent({
  setup() {
    const { app } = useContext();
    provide(
      PriorityClassStoreKey,
      priorityClassStore(app.$k8sSchedulingInstance)
    );
    return {};
  },
});
</script>
