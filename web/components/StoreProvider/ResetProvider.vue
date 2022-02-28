<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide } from "@nuxtjs/composition-api";
import resetStore from "../../store/reset";
import ResetStoreKey from "../StoreKey/ResetStoreKey";

export default defineComponent({
  setup() {
    provide(ResetStoreKey, resetStore());
    return {};
  },
});
</script>
