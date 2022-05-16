<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide, useContext } from "@nuxtjs/composition-api";
import PersistentVolumeStore from "../../store/pv";
import PersistentVolumeStoreKey from "../StoreKey/PVStoreKey";

export default defineComponent({
  setup() {
    const { app } = useContext();
    provide(PersistentVolumeStoreKey, PersistentVolumeStore(app.$k8sInstance));
    return {};
  },
});
</script>
