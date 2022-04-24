<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide, useContext } from "@nuxtjs/composition-api";
import storageClassStore from "../../store/storageclass";
import StorageClassStoreKey from "../StoreKey/StorageClassStoreKey";

export default defineComponent({
  setup() {
    const { app } = useContext();
    provide(StorageClassStoreKey, storageClassStore(app.$k8sStorageInstance));
    return {};
  },
});
</script>
