<template>
  <v-snackbar
    :value="isOpen"
    bottom
    :timeout="3000"
    :color="color()"
    @input="setIsOpen"
  >
    {{ message }}
    <v-btn color="white" text @click="closeSnackbar">Close</v-btn>
  </v-snackbar>
</template>
<script lang="ts">
import { inject, defineComponent, computed } from "@nuxtjs/composition-api";
import SnackBarStoreKey from "./StoreKey/SnackBarStoreKey";

export default defineComponent({
  setup() {
    const store = inject(SnackBarStoreKey);
    if (!store) {
      throw new Error(`${SnackBarStoreKey.description} is not provided`);
    }

    const color = () => {
      return store.messageType === "info" ? "primary" : "error";
    };
    const setIsOpen = (b: boolean) => {
      store.setIsOpen(b);
    };
    const closeSnackbar = () => store.close();

    const isOpen = computed(() => store.isOpen);
    const message = computed(() => store.message);

    return {
      color,
      setIsOpen,
      closeSnackbar,
      message,
      isOpen,
    };
  },
});
</script>
