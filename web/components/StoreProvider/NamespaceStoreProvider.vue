<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { provide, defineComponent } from "@nuxtjs/composition-api";
import namespaceStore from "../../store/namespace";
import NamespaceStoreKey from "../StoreKey/NamespaceStoreKey";

export default defineComponent({
  setup() {
    provide(NamespaceStoreKey, namespaceStore());
    return {};
  },
});
</script>
