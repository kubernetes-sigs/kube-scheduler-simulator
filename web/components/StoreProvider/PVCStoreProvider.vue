<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide } from "@nuxtjs/composition-api";
import PersistentVolumeClaimStore from "../../store/pvc";
import PersistentVolumeClaimStoreKey from "../StoreKey/PVCStoreKey";

export default defineComponent({
  setup() {
    provide(PersistentVolumeClaimStoreKey, PersistentVolumeClaimStore());
    return {};
  },
});
</script>
