<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { provide, defineComponent } from "@nuxtjs/composition-api";
import schedulerconfigurationStore from "../../store/schedulerconfiguration";
import SchedulerConfigurationStoreKey from "../StoreKey/SchedulerConfigurationStoreKey";

export default defineComponent({
  setup() {
    provide(SchedulerConfigurationStoreKey, schedulerconfigurationStore());
    return {};
  },
});
</script>
