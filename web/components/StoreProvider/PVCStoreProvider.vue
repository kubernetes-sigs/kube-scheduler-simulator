<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide, useContext } from "@nuxtjs/composition-api";
import PersistentVolumeClaimStore from "../../store/pvc";
import PersistentVolumeClaimStoreKey from "../StoreKey/PVCStoreKey";

export default defineComponent({
  setup() {
    const { app } = useContext();
    provide(
      PersistentVolumeClaimStoreKey,
      PersistentVolumeClaimStore(app.$k8sInstance)
    );
    return {};
  },
});
</script>
