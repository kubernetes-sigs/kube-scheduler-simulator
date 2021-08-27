<template>
  <v-btn class="ma-2" @click="onClick">
    <v-icon> mdi-cog </v-icon>
  </v-btn>
</template>

<script lang="ts">
import { inject, defineComponent } from "@nuxtjs/composition-api";
import SchedulerConfigurationStoreKey from "./StoreKey/SchedulerConfigurationStoreKey";

export default defineComponent({
  setup() {
    const schedulerconfigurationstore = inject(SchedulerConfigurationStoreKey);
    if (!schedulerconfigurationstore) {
      throw new Error(`${SchedulerConfigurationStoreKey} is not provided`);
    }

    const onClick = () => {
      schedulerconfigurationstore.fetchSelected();
    };
    return {
      onClick,
    };
  },
});
</script>
