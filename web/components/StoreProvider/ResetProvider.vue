<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide, useContext } from "@nuxtjs/composition-api";
import resetStore from "../../store/reset";
import ResetStoreKey from "../StoreKey/ResetStoreKey";

export default defineComponent({
  setup() {
    const { app } = useContext();
    provide(ResetStoreKey, resetStore(app.$instance));
    return {};
  },
});
</script>
