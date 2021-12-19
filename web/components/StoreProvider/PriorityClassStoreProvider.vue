<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide } from "@nuxtjs/composition-api";
import priorityClassStore from "../../store/priorityclass";
import PriorityClassStoreKey from "../StoreKey/PriorityClassStoreKey";

export default defineComponent({
  setup() {
    provide(PriorityClassStoreKey, priorityClassStore());
    return {};
  },
});
</script>
