<template>
  <v-btn class="ma-2" @click="onClick()" v-if="!enabled()">
    <v-icon> mdi-cog </v-icon>
  </v-btn>
</template>

<script lang="ts">
import { inject, defineComponent, onMounted } from "@nuxtjs/composition-api";
import SchedulerConfigurationStoreKey from "../StoreKey/SchedulerConfigurationStoreKey";

export default defineComponent({
  setup() {
    const schedulerconfigurationstore = inject(SchedulerConfigurationStoreKey);
    if (!schedulerconfigurationstore) {
      throw new Error(`${SchedulerConfigurationStoreKey.description} is not provided`);
    }

    const initializeSchedulerConfigurationStore = () => {
      schedulerconfigurationstore.initialize();
    };
    onMounted(initializeSchedulerConfigurationStore);

    const enabled = () => {
      return schedulerconfigurationstore.disabled;
    };

    const onClick = () => {
      schedulerconfigurationstore.fetchSelected();
    };
    return {
      onClick,
      enabled,
    };
  },
});
</script>
