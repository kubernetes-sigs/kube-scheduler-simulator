<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { provide, defineComponent, useContext } from "@nuxtjs/composition-api";
import schedulerconfigurationStore from "../../store/schedulerconfiguration";
import SchedulerConfigurationStoreKey from "../StoreKey/SchedulerConfigurationStoreKey";

export default defineComponent({
  setup() {
    const { app } = useContext();
    provide(
      SchedulerConfigurationStoreKey,
      schedulerconfigurationStore(app.$instance)
    );
    return {};
  },
});
</script>
