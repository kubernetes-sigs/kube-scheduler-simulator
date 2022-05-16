<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { provide, defineComponent, useContext } from "@nuxtjs/composition-api";
import nodeStore from "../../store/node";
import NodeStoreKey from "../StoreKey/NodeStoreKey";

export default defineComponent({
  setup() {
    const { app } = useContext();
    provide(NodeStoreKey, nodeStore(app.$k8sInstance));
    return {};
  },
});
</script>
