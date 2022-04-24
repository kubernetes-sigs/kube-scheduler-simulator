<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide, useContext } from "@nuxtjs/composition-api";
import podStore from "../../store/pod";
import PodStoreKey from "../StoreKey/PodStoreKey";

export default defineComponent({
  setup() {
    const { app } = useContext();
    provide(PodStoreKey, podStore(app.$k8sInstance));
    return {};
  },
});
</script>
