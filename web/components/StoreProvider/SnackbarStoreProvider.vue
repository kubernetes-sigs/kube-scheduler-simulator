<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide } from "@nuxtjs/composition-api";
import snackbarStore from "../../store/snackbar";
import SnackBarStoreKey from "../StoreKey/SnackBarStoreKey";

export default defineComponent({
  setup() {
    provide(SnackBarStoreKey, snackbarStore());
    return {};
  },
});
</script>
